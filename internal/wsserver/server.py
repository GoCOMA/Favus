from fastapi import FastAPI, WebSocket, WebSocketDisconnect, Query, Body
import boto3, os, asyncio
from typing import List, Set
from pydantic import BaseModel

app = FastAPI()

# ================== S3 =====================
s3 = boto3.client(
    "s3",
    aws_access_key_id="test",  # dummy 값
    aws_secret_access_key="test",
    region_name="ap-northeast-2",
    endpoint_url="http://localhost:4566",  # LocalStack
)

@app.get("/healthz")
def healthz():
    return {"status": "ok"}

@app.post("/create-upload")
def create_upload(bucket: str = Query(...), filename: str = Query(...)):
    resp = s3.create_multipart_upload(Bucket=bucket, Key=filename)
    return {"bucket": bucket, "key": filename, "uploadId": resp["UploadId"]}

@app.post("/presign-part")
def presign_part(
    bucket: str = Query(...),
    key: str = Query(...),
    upload_id: str = Query(...),
    part_number: int = Query(...),
    expires_in: int = Query(3600),
):
    url = s3.generate_presigned_url(
        "upload_part",
        Params={
            "Bucket": bucket,
            "Key": key,
            "UploadId": upload_id,
            "PartNumber": part_number,
        },
        ExpiresIn=expires_in,
    )
    return {"url": url, "partNumber": part_number}

class Part(BaseModel):
    ETag: str
    PartNumber: int

@app.post("/complete-upload")
def complete_upload(
    bucket: str = Query(...),
    key: str = Query(...),
    upload_id: str = Query(...),
    parts: List[Part] = Body(...),
):
    resp = s3.complete_multipart_upload(
        Bucket=bucket,
        Key=key,
        UploadId=upload_id,
        MultipartUpload={"Parts": [p.dict() for p in parts]},
    )
    return {"location": resp["Location"]}


# ================== WebSocket =====================
clients: Set[WebSocket] = set()
lock = asyncio.Lock()

async def broadcast(message: str, sender: WebSocket = None):
    """모든 연결된 클라이언트에게 메시지 전송"""
    async with lock:
        dead = []
        for client in clients:
            if client == sender:
                continue
            try:
                await client.send_text(message)
            except Exception:
                dead.append(client)
        for d in dead:
            clients.remove(d)

@app.websocket("/ws")
async def websocket_endpoint(ws: WebSocket):
    await ws.accept()
    async with lock:
        clients.add(ws)
    try:
        while True:
            data = await ws.receive_text()  # CLI에서 보내는 JSON(progress 등)
            # 모든 프론트로 broadcast
            await broadcast(data, sender=ws)
    except WebSocketDisconnect:
        async with lock:
            clients.remove(ws)
