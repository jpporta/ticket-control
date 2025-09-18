#set page(width: 300pt, height: auto)
#set align(center)
#set text(
		font: "JetBrainsMono NF",
		size: 16pt
)

#let pat = tiling(size: (30pt, 30pt))[
  #place(line(start: (0%, 0%), end: (100%, 100%)))
  #place(line(start: (0%, 100%), end: (100%, 0%)))
]
#rect(fill: pat, width: 100%, height: 60pt, stroke: 1pt)

= {{ .Day.Format "02 January 2006" }}
{{ if not .EndDay.IsZero }}
= {{ .EndDay.Format "02 January 2006" }}
{{ end }}
== {{ .CreatedBy }}

#table(
  columns: (1fr, auto),
  inset: 10pt,
  align: center,
		[N#super[o] Tasks Created],
		[{{ .NoTasks }}],
		[N#super[o] Tasks Completed],
		[{{ .NoDone }}]
)

#let pat = tiling(size: (30pt, 30pt))[
  #place(line(start: (0%, 0%), end: (100%, 100%)))
  #place(line(start: (0%, 100%), end: (100%, 0%)))
]
#rect(fill: pat, width: 100%, height: 60pt, stroke: 1pt)
