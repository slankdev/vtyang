package vtyang

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"

	vtyangapi "github.com/slankdev/vtyang/pkg/grpc/api"
	"github.com/slankdev/vtyang/pkg/liner"
	"github.com/slankdev/vtyang/pkg/util"
)

var (
	GlobalOptRunFilePath string

	actionCBs = map[string]func(args []string) error{
		"uptime-callback": func(args []string) error {
			fmt.Fprintf(stdout, "UPTIME")
			return nil
		},
		"date-callback": func(args []string) error {
			fmt.Fprint(stdout, "DATE")
			return nil
		},
	}
	_ = actionCBs

	exit            bool      = false
	stdout          io.Writer = os.Stdout
	cliMode         CliMode   = CliModeView
	dbm             *DatabaseManager
	commitHistories []CommitHistory
	commandnodes    map[CliMode]*CommandNode
	yangmodules     *yang.Modules
)

const (
	QUESTION_MARK rune = 63 // '?'
)

func getDatabasePath() string {
	return fmt.Sprintf("%s/config.json", GlobalOptRunFilePath)
}

func getPrompt() string {
	switch cliMode {
	case CliModeView:
		return "vtyang# "
	case CliModeConfigure:
		return "vtyang(config)# "
	default:
		panic(fmt.Sprintf("CLIMODE(%v)", cliMode))
	}
}

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "vtyang",
	}
	rootCmd.AddCommand(util.NewCommandCompletion(rootCmd))
	rootCmd.AddCommand(util.NewCommandVersion())
	rootCmd.AddCommand(newCommandAgent())
	rootCmd.AddCommand(newCommandWatch())
	return rootCmd
}

func InitAgent(runtimePath, yangPath string) error {
	if runtimePath != "" {
		if err := os.MkdirAll(runtimePath, 0777); err != nil {
			return err
		}
	}

	GlobalOptRunFilePath = runtimePath
	dbm = NewDatabaseManager()
	dbm.LoadYangModuleOrDie(yangPath)
	if err := dbm.LoadDatabaseFromFile(getDatabasePath()); err != nil {
		return err
	}

	var err error
	yangmodules, err = yangModulesPath(yangPath)
	if err != nil {
		return err
	}

	cliMode = CliModeView
	installCommandsDefault(CliModeView)
	installCommandsDefault(CliModeConfigure)
	installCommands()
	initCommitHistories()
	installCommandsPostProcess()

	if GlobalOptRunFilePath != "" {
		if err := os.MkdirAll(GlobalOptRunFilePath, 0777); err != nil {
			return err
		}
	}
	return nil
}

func newCommandWatch() *cobra.Command {
	cmd := &cobra.Command{
		Use: "watch",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := grpc.Dial(
				"localhost:8080",
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithBlock(),
			)
			if err != nil {
				return err
			}
			defer conn.Close()
			client := vtyangapi.NewGreetingServiceClient(conn)
			stream, err := client.HelloStream(context.Background())
			if err != nil {
				return err
			}
			loop := true
			for loop {
				res, err := stream.Recv()
				if err != nil {
					fmt.Println(err.Error())
					loop = false
					continue
				}
				m := map[string]interface{}{}
				if err := json.Unmarshal([]byte(res.Data), &m); err != nil {
					return err
				}
				pp.Println(m)
			}

			return nil
		},
	}
	return cmd
}

func newCommandAgent() *cobra.Command {
	cmd := &cobra.Command{
		Use: "agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			// XXX
			if err := InitAgent(GlobalOptRunFilePath, "./yang"); err != nil {
				return err
			}

			line := liner.NewLiner()
			defer line.Close()
			line.SetCtrlCAborts(true)
			line.SetWordCompleter(completer)
			line.SetTabCompletionStyle(liner.TabPrints)
			line.SetBinder(QUESTION_MARK, completionLister)

			go startRPCServer()
			for {
				if name, err := line.Prompt(getPrompt()); err == nil {
					line.AppendHistory(name)
					name = strings.TrimSpace(name)
					args := strings.Fields(name)
					if len(args) == 0 {
						continue
					}
					cn := getCommandNodeCurrent()
					cn.executeCommand(cat(args))
				} else if err == liner.ErrPromptAborted {
					log.Print("aborted")
					break
				} else {
					log.Print("error reading line: ", err)
				}
				if exit {
					break
				}
			}
			return nil
		},
	}
	fs := cmd.Flags()
	fs.StringVarP(&GlobalOptRunFilePath, "run-path", "r", "", "Runtime file path")
	return cmd
}

func init() {
	logfile, err := os.OpenFile("/tmp/vtyang.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	log.SetOutput(logfile)
	log.Printf("starting vtyang...\n")
}

func ErrorOnDie(err error) {
	if err != nil {
		panic(err)
	}
}

func startRPCServer() {
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	vtyangapi.RegisterGreetingServiceServer(s, &myServer{})
	reflection.Register(s)
	s.Serve(listener)
	// TODO(slankdev): terminating process
}

type myServer struct {
	vtyangapi.UnimplementedGreetingServiceServer
}

func (s *myServer) HelloStream(
	stream vtyangapi.GreetingService_HelloStreamServer) error {
	// Init loop stopper
	loop := true
	go func() {
		for loop {
			if _, err := stream.Recv(); err != nil {
				if err == io.EOF {
					pp.Println("EOF")
				} else {
					pp.Println(err.Error())
				}
				loop = false
			}
		}
	}()

	// Once Flush current running-config
	xpath, _, err := ParseXPathArgs(dbm, []string{}, false)
	if err != nil {
		return err
	}
	node, err := dbm.GetNode(xpath)
	if err != nil {
		return err
	}
	if err := stream.Send(&vtyangapi.HelloResponse{
		Data: node.String(),
	}); err != nil {
		return err
	}

	// Get remote ip
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return fmt.Errorf("peer.FromContext is not ok")
	}

	// Watch Configuration change and notify it
	confChan := make(chan Configuration)
	confChans[p.Addr.String()] = confChan
	pp.Println("Starting subscribe")
	for loop {
		conf := <-confChan
		if err := stream.Send(&vtyangapi.HelloResponse{
			Message: "HELLO",
			Data:    conf.Data,
		}); err != nil {
			pp.Println(err.Error())
			loop = false
		}
	}
	pp.Println("Stopping subscribe")
	delete(confChans, p.Addr.String())
	return nil
}

type Configuration struct {
	Revision int
	Data     string
}

var (
	confChans = map[string]chan Configuration{}
)

func nofityRunningConfigToSubscribers() error {
	xpath, _, err := ParseXPathArgs(dbm, []string{}, false)
	if err != nil {
		return err
	}
	node, err := dbm.GetNode(xpath)
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err.Error())
		return err
	}
	fmt.Fprintln(stdout, node.String())
	for _, confChan := range confChans {
		confChan <- Configuration{
			Data: node.String(),
		}
	}
	return nil
}
