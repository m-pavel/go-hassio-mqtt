package ghm

type IndexModel struct {
	Yaxis string
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

function step() {
	fetch("/api/v1/data")
		.then(response => response.json())
		.then(json => {
			for (let e of json) {
				e.ts = new Date(e.ts * 1000);
			}
			var mark = Plot.line(json, {x: "ts", y: "value", stroke: "#fc7303", curve: "basis"});
			var options = {
				y: {
					grid: true,
					label: "{{ .Yaxis }}"
				},
				marks: [
					mark
				],
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
