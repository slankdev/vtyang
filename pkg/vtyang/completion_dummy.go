package vtyang

var tree = CompletionTree{
	Root: CompletionNode{
		Name:        "",
		Description: "",
		Level:       -1,
		Childs: []CompletionNode{
			{
				Name:        "test",
				Description: "Test command",
				Level:       0,
				Childs: []CompletionNode{
					{
						Name: "NAME",
						Childs: []CompletionNode{
							{
								Name: "age",
								Childs: []CompletionNode{
									{
										Name:   "NAME",
										Childs: []CompletionNode{{Name: "<cr>"}},
									},
								},
							},
							{
								Name: "mail",
								Childs: []CompletionNode{
									{
										Name:   "NAME",
										Childs: []CompletionNode{{Name: "<cr>"}},
									},
								},
							},
						},
					},
				},
			},
			{
				Name:        "set",
				Description: "Set system parameter",
				Level:       0,
			},
			{
				Name:        "show",
				Description: "Display information",
				Level:       0,
				Childs: []CompletionNode{
					{
						Name:        "running-config",
						Description: "Display current configuration",
						Level:       1,
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "startup-config",
						Description: "Display startup configuration",
						Level:       1,
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "operational-data",
						Description: "Display operational data",
						Level:       1,
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "commit",
						Description: "Display commit information",
						Level:       1,
						Childs: []CompletionNode{
							{
								Name:        "history",
								Description: "Display commit history",
								Level:       2,
								Childs:      []CompletionNode{{Name: "<cr>"}},
							},
						},
					},
					{
						Name:        "yang-modules",
						Description: "Display yang modules",
						Level:       1,
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
					{
						Name:        "cli-tree",
						Description: "Display cli tree dump",
						Level:       1,
						Childs:      []CompletionNode{{Name: "<cr>"}},
					},
				},
			},
			{
				Name:        "quit",
				Description: "Quit system",
				Level:       0,
				Childs:      []CompletionNode{{Name: "<cr>"}},
			},
		},
	},
}
