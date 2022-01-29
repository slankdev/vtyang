package vtyang

var viewTree = CompletionTree{
	Root: CompletionNode{
		Name:        "",
		Description: "",
		Childs: []CompletionNode{
			{
				Name:        "configure",
				Description: "Enable swconfigure mode",
				Childs:      []CompletionNode{{Name: "<cr>"}},
			},
			{
				Name:        "write",
				Description: "Write system parameter",
				Childs: []CompletionNode{
					{
						Name:        "memory",
						Description: "Write system parameter to memory",
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
				},
			},
			{
				Name:        "show",
				Description: "Display information",
				Childs: []CompletionNode{
					{
						Name:        "running-config",
						Description: "Display current configuration",
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "startup-config",
						Description: "Display startup configuration",
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "operational-data",
						Description: "Display operational data",
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "commit",
						Description: "Display commit information",
						Childs: []CompletionNode{
							{
								Name:        "history",
								Description: "Display commit history",
								Childs:      []CompletionNode{{Name: "<cr>"}},
							},
						},
					},
					{
						Name:        "yang-modules",
						Description: "Display yang modules",
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "cli-tree",
						Description: "Display cli tree dump",
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "database-tree",
						Description: "Display database dump",
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
				},
			},
			{
				Name:        "quit",
				Description: "Quit system",
				Childs:      []CompletionNode{{Name: "<cr>"}},
			},
		},
	},
}
var configureTree = CompletionTree{
	Root: CompletionNode{
		Name:        "",
		Description: "",
		Childs: []CompletionNode{
			{
				Name:        "delete",
				Description: "Delete system parameter",
			},
			{
				Name:        "set",
				Description: "Set system parameter",
			},
		},
	},
}
