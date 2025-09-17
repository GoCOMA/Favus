import asyncio
import websockets
import logging

logging.basicConfig(level=logging.INFO)

CONNECTED_CLIENTS = set()

async def handler(connection, path=None):
    CONNECTED_CLIENTS.add(connection)
    print(f"✅ Client connected: {connection.remote_address}")

    try:
        async for message in connection:
            print(f"📩 Received: {message}")
            # 브로드캐스트 (보낸 사람 제외, 열린 연결만)
            for client in CONNECTED_CLIENTS:
                if client != connection and not client.closed:
                    try:
                        await client.send(message)
                    except websockets.ConnectionClosed as e:
                        print(f"⚠️ Send failed to {client.remote_address}: {e}")
                    except Exception as e:
                        print(f"⚠️ Send failed with other error to {client.remote_address}: {e}")

    except websockets.ConnectionClosed:
        print(f"❌ Client disconnected: {connection.remote_address}")
    finally:
        CONNECTED_CLIENTS.discard(connection)

async def main():
    async with websockets.serve(
        handler,
        "0.0.0.0",
        8765,
    ):
        logging.info("🚀 WebSocket server running at ws://0.0.0.0:8765/ws")
        await asyncio.Future()  # 서버 계속 실행

if __name__ == "__main__":
    asyncio.run(main())
