# openserverless-streamer

The openserverless streamer is a tool to relay a stream from OpenWhisk actions to an outside
HTTP client.

The streamer is a simple HTTP server that exposes an endpoint /stream/{namespace}/{action} to 
invoke the relative OpenWhisk action, open a socket for the action to write to, and relay the
output to the client.

It expects 2 environment variables to be set:
- `APIHOST`: the OpenWhisk API host
- `STREAMER_ADDR`: the address of the streamer server for the OpenWhisk actions to connect to

