const express = require('express');
const app = express();
const fs = require('fs');
const httpProxy = require('http-proxy');
const apiProxy = httpProxy.createProxyServer();

app.all('*', function(req, res) {
    const routesFile = readJsonFile('routes.json');
    const servers = routesFile['servers'];
    const routes = routesFile['routes'];

    let urlThatMatters = getPathFromUrl(req.originalUrl);
    let serverName = routes[urlThatMatters]["server"];
    let site = servers[serverName]["site"];

    res.setHeader('Content-Type', 'application/json');
    apiProxy.web(req, res, {
        target: site + urlThatMatters,
        changeOrigin: true,
        ignorePath: true,
    });
});

app.listen(3000, () => console.log('API is listening on port 3000!'));

function readJsonFile(filepath, encoding){
    if (typeof (encoding) == 'undefined') {
        encoding = 'utf8';
    }
    var file = fs.readFileSync(filepath, encoding);
    return JSON.parse(file);
}

// it should never happen, but if the path doesn't match, this function will return false
function getPathFromUrl(url) {
    let regex = new RegExp('^/api(.*)');

    // match[1] has the group that was used in regex
    return ((match = regex.exec(url)) !== null && match[1] != '' ? match[1] : false);
}
