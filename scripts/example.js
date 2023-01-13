// ==BocchiScript==
// @name         Test
// @version      0.1
// @description  This is a example script
// @author       WorldObservationLog
// @match        https://httpbin.org:443/post
// @match        https://httpbin.org:443/get
// @priority     100
// ==/BocchiScript==

function OnResponse(resp) {
    return resp
}


function OnRequest(req) {
    return req
}


