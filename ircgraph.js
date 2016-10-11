if (typeof graphsource === 'undefined') {
    graphsource = "map.json"
}

var svg = d3.select("svg"), width = +svg.attr("width"),
    height = +svg.attr("height");

var color = d3.scaleOrdinal(d3.schemeCategory20);

var simulation =
    d3.forceSimulation()
        .force("link", d3.forceLink().id(function(d) { return d.id; }))
        .force("charge", d3.forceManyBody())
        .force("center", d3.forceCenter(width / 2, height / 2));

d3.json(graphsource, function(error, graph) {
    if (error) throw error;

    var link =
        svg.append("g")
            .attr("class", "links")
            .selectAll("line")
            .data(graph.Links)
            .enter()
            .append("line")
            .attr("stroke-width", function(d) { return Math.sqrt(d.lag); });

    var node = svg.append("g")
                   .attr("class", "nodes")
                   .selectAll("circle")
                   .data(graph.Servers)
                   .enter()
                   .append("g")
                   .call(
                       d3.drag()
                           .on("start", dragstarted)
                           .on("drag", dragged)
                           .on("end", dragended));

    node.append("circle")
        .attr("r", function(d) { return d.usercount/10 + 3; })
        .attr("fill", function(d) { return color(d.group); });

    node.append("title").text(function(d) { return d.desc + " (" + d.usercount + " Utilisateurs)"; });

    node.append("text")
        .attr("dx", function(d) { return d.usercount/10 + 8; })
        .attr("dy", ".35em")
        .text(function(d) { return d.label; });

    simulation.nodes(graph.Servers).on("tick", ticked);

    simulation.force("link").links(graph.Links).distance(function(d) {
        return 50 + Math.sqrt(d.lag) * 10;
    });

    function ticked() {
        link.attr("x1", function(d) { return d.source.x; })
            .attr("y1", function(d) { return d.source.y; })
            .attr("x2", function(d) { return d.target.x; })
            .attr("y2", function(d) { return d.target.y; });

        node.selectAll("circle")
            .attr("cx", function(d) { return d.x; })
            .attr("cy", function(d) { return d.y; });
        node.selectAll("text")
            .attr("x", function(d) { return d.x; })
            .attr("y", function(d) { return d.y; });
    }
});

function dragstarted(d) {
    if (!d3.event.active) simulation.alphaTarget(0.3).restart();
    d.fx = d.x;
    d.fy = d.y;
}

function dragged(d) {
    d.fx = d3.event.x;
    d.fy = d3.event.y;
}

function dragended(d) {
    if (!d3.event.active) simulation.alphaTarget(0);
    d.fx = null;
    d.fy = null;
}
