#set page(
  width: 300pt,
  height: auto,
  margin: (x: 14pt, top: 16pt, bottom: 20pt)
)
#set text(
  font: (
    {{ if .Font }}"{{ .Font }}",{{ end }}
    "Libertinus Serif",
    "DejaVu Serif",
    "BerkeleyMono Nerd Font Mono",
    "Liberation Serif"
  ),
  size: {{ if .FontSize }}{{ .FontSize }}{{ else }}11pt{{ end }}
)
#set par(
  justify: {{ if .Justify }}true{{ else }}false{{ end }},
  leading: 0.65em
)

{{ if .Title }}
#align(center)[
  #text(size: 14pt, weight: "bold")[{{ .Title }}]
  {{ if .Date }}
  #v(2pt)
  #text(size: 9pt, fill: luma(80))[{{ .Date }}]
  {{ end }}
  #v(4pt)
  #line(length: 60%, stroke: 0.8pt)
]
#v(6pt)
{{ else if .Date }}
#align(right)[
  #text(size: 9pt, fill: luma(80))[{{ .Date }}]
]
#v(4pt)
{{ end }}

{{ if .To }}
*{{ if .ToLabel }}{{ .ToLabel }}{{ else }}Para{{ end }}:* {{ .To }}\
#v(4pt)
{{ end }}

{{ .Content }}

{{ if .From }}
#v(12pt)
#align(right)[
  {{ if .SignOff }}{{ .SignOff }}\{{ end }}
  #v(2pt)
  *{{ .From }}*
]
{{ end }}
