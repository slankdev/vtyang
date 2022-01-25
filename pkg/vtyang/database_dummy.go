package vtyang

// set account users user hiroki age 26
// set account users user hiroki projects tennis
// set account users user hiroki projects driving
// set account users user slankdev age 99
// set account users user slankdev projects kloudnfv
// set account users user slankdev projects wide

var DummyDBRoot = DBNode{
	Name: "<root>",
	Type: Container,
	Childs: []DBNode{
		{
			Name: "users",
			Type: Container,
			Childs: []DBNode{
				{
					Name: "user",
					Type: List,
					Childs: []DBNode{
						{
							Name: ".",
							Type: Container,
							Childs: []DBNode{
								{
									Name: "name",
									Type: Leaf,
									Value: DBValue{
										Type:   YString,
										String: "hiroki",
									},
								},
								{
									Name: "age",
									Type: Leaf,
									Value: DBValue{
										Type:    YInteger,
										Integer: 26,
									},
								},
								{
									Name: "projects",
									Type: List,
									Childs: []DBNode{
										{
											Name: ".",
											Type: Container,
											Childs: []DBNode{
												{
													Name: "name",
													Type: Leaf,
													Value: DBValue{
														Type:   YString,
														String: "tennis",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:   YBoolean,
														String: "true",
													},
												},
											},
										},
										{
											Name: ".",
											Type: Container,
											Childs: []DBNode{
												{
													Name: "name",
													Type: Leaf,
													Value: DBValue{
														Type:   YString,
														String: "driving",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:   YBoolean,
														String: "false",
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Name: ".",
							Type: Container,
							Childs: []DBNode{
								{
									Name: "name",
									Type: Leaf,
									Value: DBValue{
										Type:   YString,
										String: "slankdev",
									},
								},
								{
									Name: "age",
									Type: Leaf,
									Value: DBValue{
										Type:    YInteger,
										Integer: 36,
									},
								},
								{
									Name: "projects",
									Type: List,
									Childs: []DBNode{
										{
											Name: ".",
											Type: Container,
											Childs: []DBNode{
												{
													Name: "name",
													Type: Leaf,
													Value: DBValue{
														Type:   YString,
														String: "kloudnfv",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:   YBoolean,
														String: "false",
													},
												},
											},
										},
										{
											Name: ".",
											Type: Container,
											Childs: []DBNode{
												{
													Name: "name",
													Type: Leaf,
													Value: DBValue{
														Type:   YString,
														String: "wide",
													},
												},
												{
													Name: "finished",
													Type: Leaf,
													Value: DBValue{
														Type:   YBoolean,
														String: "false",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
