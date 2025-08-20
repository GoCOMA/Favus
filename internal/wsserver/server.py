import asyncio
import websockets
import logging

logging.basicConfig(level=logging.INFO)

CONNECTED_CLIENTS = set()

async def handler(connection: websockets.ServerConnection):
    CONNECTED_CLIENTS.add(connection)
    print(f"✅ Client connected: {connection.remote_address}")

    try:
        async for message in connection:
            print(f"📩 Received: {message}")
            # 브로드캐스트 (보낸 사람 제외, 열린 연결만)
            for client in CONNECTED_CLIENTS:
                if client != connection and client.state.name == "OPEN":
                    try:
                        await client.send(message)
                        # print(f"📤 Sent to {client.remote_address}: {message}")
                    except websockets.ConnectionClosed as e:
                        print(f"⚠️ Send failed to {client.remote_address}: {e}")
                    except Exception as e:
                        print(f"⚠️ Send failed with other error to {client.remote_address}: {e}")

    except websockets.ConnectionClosed:
        print(f"❌ Client disconnected: {connection.remote_address}")
    finally:
        CONNECTED_CLIENTS.remove(connection)


async def process_request(path, request_headers):
    logging.info(f"Request headers: {request_headers}")
    return None  # 계속 진행

async def main():
    async with websockets.serve(
        handler,
        "127.0.0.1",
        8765,
        process_request=process_request,
        origins=["http://localhost:3000", None],
        ping_interval=None, # 클라이언트 ping에만 의존
    ):
        logging.info("🚀 WebSocket server running at ws://127.0.0.1:8765/ws")
        await asyncio.Future()  # 서버 계속 실행

if __name__ == "__main__":
    asyncio.run(main())
