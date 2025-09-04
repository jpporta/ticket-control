#set page(width: 300pt, height: auto)
#set align(center)
#set text(
		font: "JetBrainsMono NF",
		size: 13pt
)

{{ .CreatedAt.Format "2006-01-02 15:04" }}

[ {{ .ID }} ]

{{ .PriorityDisplay }}

= {{ .Title }}
#sub[{{ .CreatedBy }}]
#line(length: 50%)


{{ .Description }}
