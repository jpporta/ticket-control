#set page(width: 300pt, height: auto)
#set align(center)
#set text(
		font: "BerkeleyMono Nerd Font",
		size: 13pt,
)
= {{ .Title }}
#sub[{{ .CreatedBy }}]
#line(length: 50%)

#set align(left)
{{ range .Content }}
- [ ] {{ . }}
{{ end }}
