var w = 1280 - 80,
    h = 800 - 180,
    x = d3.scale.linear().range([0, w]),
    y = d3.scale.linear().range([0, h]),
    color = d3.scale.category20c(),
    root,
    node;

var treemap = d3.layout.treemap()
        .round(false)
        .size([w, h])
        .sticky(true)
        .value(function(d) { return d.cpu; });

var svg = d3.select("#body").append("div")
        .attr("class", "chart")
        .style("width", w + "px")
        .style("height", h + "px")
        .append("svg:svg")
        .attr("width", w)
        .attr("height", h)
        .append("svg:g")
        .attr("transform", "translate(.5,.5)");

var valuefns = {
    "cpu": (d => d.cpu),
    "soft-mem": (d => d.soft_memory || d.max_memory),
    "max-mem": (d => d.name == "Unused" ? 0 : d.max_memory) // Hide Unused node
}

var cluster = window.location.search.substr(1)

d3.json("/resources/" + cluster, function(data) {
    node = root = data;

    var nodes = treemap.nodes(root)
            .filter(function(d) { return !d.children; });

    var cell = svg.selectAll("g")
            .data(nodes)
            .enter().append("svg:g")
            .attr("class", "cell")
            .attr("transform", function(d) { return "translate(" + d.x + "," + d.y + ")"; })
            .on("click", function(d) { return zoom(node == d.parent ? root : d.parent); });

    cell.append("svg:rect")
        .attr("width", function(d) { return Math.max(0, d.dx - 1); })
        .attr("height", function(d) { return Math.max(0, d.dy - 1); })
        .style("fill", function(d) { return d.name == "Unused" ? d3.rgb(230,230,230) : color(d.parent.name); });

    cell.append("svg:text")
        .attr("x", function(d) { return d.dx / 2; })
        .attr("y", function(d) { return d.dy / 2; })
        .attr("dy", ".35em")
        .attr("text-anchor", "middle")
        .text(function(d) { return d.name.split(".")[0]; })
        .style("opacity", function(d) { d.w = this.getComputedTextLength(); return d.dx > d.w ? 1 : 0; });

    d3.select(window).on("click", function() { zoom(root); });

    d3.select("select").on("change", function() {
        treemap.value(valuefns[this.value]).nodes(root);
        zoom(node);
    });
});

function zoom(d) {
    var kx = w / d.dx, ky = h / d.dy;
    x.domain([d.x, d.x + d.dx]);
    y.domain([d.y, d.y + d.dy]);

    var t = svg.selectAll("g.cell").transition()
            .duration(d3.event.altKey ? 7500 : 750)
            .attr("transform", function(d) { return "translate(" + x(d.x) + "," + y(d.y) + ")"; });

    t.select("rect")
        .attr("width", function(d) { return Math.max(0, kx * d.dx - 1); })
        .attr("height", function(d) { return Math.max(0, ky * d.dy - 1); });

    t.select("text")
        .attr("x", function(d) { return kx * d.dx / 2; })
        .attr("y", function(d) { return ky * d.dy / 2; })
        .style("opacity", function(d) { return kx * d.dx > d.w ? 1 : 0; });

    node = d;
    d3.event.stopPropagation();
}
