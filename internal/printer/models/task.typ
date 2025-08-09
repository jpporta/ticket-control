#set page(width: 300pt, height: auto)
#set align(center)
#set text(
		font: "JetBrainsMono NF",
)

{{ .PriorityDisplay }}
= {{ .Title }}
#sub[{{ .CreatedBy }}]
#line(length: 50%)


{{ .Description }}
