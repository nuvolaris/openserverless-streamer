import socket
import os
import time

example_data = [
    "Hello, World!",
    "This is a test",
    "from an HTTP SSE request",
    "Through an openwhisk action",
    "To a socket server",
    "Back to the HTTP client",
]

def main(args):

    streamer = (args.get("STREAM_HOST"), args.get("STREAM_PORT"))
    
    if not streamer[0] or not streamer[1]:
        return {"body": "please provide a STREAM_HOST and STREAM_PORT"}
    
    print(f"streamer: {streamer}")

    # # invoke a call to a streaming api server like OpenAI
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.connect((streamer[0], int(streamer[1])))

        for ex in example_data:
            time.sleep(1)
            if ex:
                # send data
                print(f"sending: {ex}")
                s.sendall(ex.encode())

        print("done sending. closing connection")
        s.sendall("EOF".encode())
        s.close()
    
    return {"body": "done"}