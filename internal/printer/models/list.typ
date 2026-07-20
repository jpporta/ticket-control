#set page(width: 300pt, height: auto)
#set text(
	font: "BerkeleyMono Nerd Font Mono",
	size: 13pt,
)
#set par(leading: 0.55em)
{{ if .Title }}
#set align(center)
= {{ .Title }}
#sub[{{ .CreatedBy }}]
#block(below: 12pt, line(length: 50%))
{{ end }}
#set align(left)
#table(
	columns: (30%, 70%),
	inset: (top: 8pt, right: 6pt, bottom: 1.1em, left: 6pt),
	stroke: (x, y) => (
		top: if y > 0 { black + 0.6pt },
		left: if x > 0 { black + 0.6pt },
	),
{{ range .Content }}
	[], [{{ . }}],
{{ end }}
)
