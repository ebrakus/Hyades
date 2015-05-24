var http=require('http');
var ports = [];

for(var i = 0; i < 6; i++) {
 ports.push(9000+i)
 ports.push(9100+i)
 ports.push(9010+i)
 ports.push(9110+i)
}

var servers = [];
var s;
function reqHandler(req, res) {
    console.log({
        remoteAddress: req.socket.remoteAddress,
        remotePort: req.socket.remotePort,
        localAddress: req.socket.localAddress,
        localPort: req.socket.localPort,
    });
    res.writeHead(200, {"Content-Type": "text/html"});
    var temp;
    res.write('Response from ' + req.socket.localPort + '\n');
    res.end();
}
ports.forEach(function(port) {
    s = http.createServer(reqHandler);
    s.listen(port);
    servers.push(s);
});