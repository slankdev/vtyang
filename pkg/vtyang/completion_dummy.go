package vtyang

var viewTree = CompletionTree{
	Root: CompletionNode{
		Name:        "",
		Description: "",
		Childs: []CompletionNode{
			{
				Name:        "show",
				Description: "Display information",
				Childs: []CompletionNode{
					{
						Name:        "startup-config",
						Description: "Display startup configuration",
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
