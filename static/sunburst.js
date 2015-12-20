var width = 960,
    height = 900,
    radius = Math.min(width, height) / 2 - 30;

var x = d3.scale.linear()
        .range([0, 2 * Math.PI]);

var y = d3.scale.sqrt()
        .range([0, radius]);

var color = d3.scale.category20();

var colorByApp = function(d) {
    return d.children ? utilColor(d) : color(d.name.split(".")[0]);
};

var colorByEnv = function(d) {
    return d.children ? utilColor(d) : color(d.name.split(".")[1]);
};

var svg = d3.select("body").append("svg")
        .attr("width", width)
        .attr("height", height)
        .append("g")
        .attr("transform", "translate(" + width / 2 + "," + (height / 2 + 10) + ")");

var partition = d3.layout.partition()
        .value(function(d) { return d.memory; })
        .sort(null);

var tooltip = d3.select("body")
        .append("div")
        .attr("class", "tooltip");

tooltip.append('div').attr('class', 'label');
tooltip.append('div').attr('class', 'count');
tooltip.append('div').attr('class', 'percent');

function getReadableSize(sizeInBytes) {
    var i = -1;
    var byteUnits = [' GB', ' TB', 'PB', 'EB', 'ZB', 'YB'];
    do {
        sizeInBytes = sizeInBytes / 1024;
        i++;
    } while (sizeInBytes > 1024);

    return Math.max(sizeInBytes, 0.1).toFixed(1) + byteUnits[i];
};

var arc = d3.svg.arc()
        .startAngle(function(d) { return Math.max(0, Math.min(2 * Math.PI, x(d.x))); })
        .endAngle(function(d) { return Math.max(0, Math.min(2 * Math.PI, x(d.x + d.dx))); })
        .innerRadius(function(d) { return Math.max(0, y(d.y)); })
        .outerRadius(function(d) { return Math.max(0, y(d.y + d.dy)); });

// Color slave nodes based on utilization (green good, red bad)
function utilColor(d) {
    var util = (displayType == "cpu" ? d.cpu / d.cpu_total : d.memory / d.memory_total);
    var hue = (util * 110);
    return d3.hsl(hue, .9, .5);
};


// Keep track of the node that is currently being displayed as the root.
var node;

var displayType = "memory";
var colorType = "app";

d3.json("resources.json", function(error, root) {
    node = root;
    var path = svg.datum(root).selectAll("path")
            .data(partition.nodes)
            .enter().append("path")
            .attr("display", function(d) { return d.name == "Unused" ? "none" : null; })
            .attr("d", arc)
            .style("fill", colorType == "app" ? colorByApp : colorByEnv)
            .on("click", click)
            .on("mouseover", function (d) {
                var value = displayType == "cpu" ? d.cpu.toFixed(2) : d.memory;
                var total = displayType == "cpu" ? d.cpu_total : d.memory_total;
                var value_formatted = displayType == "cpu" ? d.cpu.toFixed(2) + " CPU(s)" : getReadableSize(d.memory);
                var total_formatted = displayType == "cpu" ? ((d.cpu_total && d.cpu_total.toFixed(2)) || "") + " CPU(s)" : getReadableSize(d.memory_total);
                var percent = Math.round(1000 * value / total) / 10;
                tooltip.select('.label').html(d.name/*.split(".")[0]*/);
                tooltip.select('.count').html(d.children ? value_formatted + " / " + total_formatted : value_formatted);
                tooltip.select('.percent').html(d.children ? percent + "% Utilization" : "");
                tooltip.style('display', 'block');
            })
            .on("mouseout", function() {
                tooltip.style('display', 'none');
            })
            .each(stash);

    d3.selectAll("input.value").on("change", function change() {
        var value = this.value === "cpu"
                ? function(d) { displayType = "cpu"; return d.cpu; }
            : function(d) { displayType = "memory"; return d.memory; };

        path
            .data(partition.value(value).nodes)

            .transition()
            .duration(1000)
            .style("fill", colorType === "app" ? colorByApp : colorByEnv)
            .attrTween("d", arcTweenData);
    });

    d3.selectAll("input.color").on("change", function change() {
        var colorFunc = this.value === "app"
                ? colorByApp : colorByEnv;
        colorType = this.value;

        path
            .data(partition.nodes)
            .transition()
            .duration(1000)
            .style("fill", colorFunc)
            .attrTween("d", arcTweenData);
    });

    function click(d) {
        node = d;
        path.transition()
            .duration(1000)
            .attrTween("d", arcTweenZoom(d));
    }
});

d3.select(self.frameElement).style("height", height + "px");

// Setup for switching data: stash the old values for transition.
function stash(d) {
    d.x0 = d.x;
    d.dx0 = d.dx;
}

// When switching data: interpolate the arcs in data space.
function arcTweenData(a, i) {
    var oi = d3.interpolate({x: a.x0, dx: a.dx0}, a);
    function tween(t) {
        var b = oi(t);
        a.x0 = b.x;
        a.dx0 = b.dx;
        return arc(b);
    }
    if (i == 0) {
        // If we are on the first arc, adjust the x domain to match the root node
        // at the current zoom level. (We only need to do this once.)
        var xd = d3.interpolate(x.domain(), [node.x, node.x + node.dx]);
        return function(t) {
            x.domain(xd(t));
            return tween(t);
        };
    } else {
        return tween;
    }
}

// When zooming: interpolate the scales.
function arcTweenZoom(d) {
    var xd = d3.interpolate(x.domain(), [d.x, d.x + d.dx]),
        yd = d3.interpolate(y.domain(), [d.y, 1]),
        yr = d3.interpolate(y.range(), [d.y ? 20 : 0, radius]);
    return function(d, i) {
        return i
            ? function(t) { return arc(d); }
        : function(t) { x.domain(xd(t)); y.domain(yd(t)).range(yr(t)); return arc(d); };
    };
}
