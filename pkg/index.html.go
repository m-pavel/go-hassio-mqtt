package ghm

type IndexModel struct {
	Yaxis string
	Marks []string
}

const index_html = `
<html><head>
<style>
        #svg {
            height: 100vw;
			height: 100%;
            width: 100%;
   			width: 100vw;
        }
    </style>

<script type="module">

import * as Plot from "https://cdn.jsdelivr.net/npm/@observablehq/plot@0.6/+esm";

setTimeout(step, 1000);
	var colors = [
    "#FFB300",
    "#803E75",
    "#FF6800",
    "#A6BDD7",
    "#C10020",
    "#CEA262",
    "#817066",
    "#007D34",
    "#F6768E",
    "#00538A",
    "#FF7A5C",
    "#53377A",
    "#FF8E00",
    "#B32851",
    "#F4C800",
    "#7F180D",
    "#93AA00",
    "#593315",
    "#F13A13",
    "#232C16",
    ];

function step() {
	fetch("/api/v1/data")
		.then(response => response.json())
		.then(json => {
			for (let e of json) {
				e.ts = new Date(e.ts * 1000);
			}
			var marks = [];
			{{range $i, $val := .Marks }}
				marks.push(Plot.line(json, {x: "ts", y: "{{ $val }}", z : null, stroke: colors[{{ $i }}], curve: "basis"}));
			{{end}}
			var options = {
				y: {
					grid: true,
					label: "{{ .Yaxis }}"
				},
				marks: marks,
			};
			var plot = Plot.plot(options);
			plot.id = "svg";
			document.getElementById('svg').replaceWith(plot);
		});
				
    setTimeout(step, 10000);
}

</script>
</head>
<body>
	<div id="svg" />
</bodY>
</html>
`
