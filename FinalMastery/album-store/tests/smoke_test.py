import os
import time
import uuid
import requests

BASE_URL = os.getenv("BASE_URL", "http://localhost")

def main():
    album_id = str(uuid.uuid4())

    r = requests.get(f"{BASE_URL}/health")
    print("health:", r.status_code, r.text)
    assert r.status_code == 200
    assert r.json()["status"] == "ok"

    body = {
        "album_id": album_id,
        "title": "My Summer Trip",
        "description": "Photos from Cancun",
        "owner": "student@northeastern.edu"
    }

    r = requests.put(f"{BASE_URL}/albums/{album_id}", json=body)
    print("put album:", r.status_code, r.text)
    assert r.status_code in (200, 201)

    r = requests.get(f"{BASE_URL}/albums/{album_id}")
    print("get album:", r.status_code, r.text)
    assert r.status_code == 200

    with open("sample.jpg", "rb") as f:
        r = requests.post(
            f"{BASE_URL}/albums/{album_id}/photos",
            files={"photo": ("sample.jpg", f, "image/jpeg")}
        )
    print("upload photo:", r.status_code, r.text)
    assert r.status_code == 202
    data = r.json()
    photo_id = data["photo_id"]

    for _ in range(30):
        r = requests.get(f"{BASE_URL}/albums/{album_id}/photos/{photo_id}")
        print("photo status:", r.status_code, r.text)
        assert r.status_code == 200
        pdata = r.json()
        if pdata["status"] == "completed":
            url = pdata["url"]
            rr = requests.get(url)
            print("photo url:", rr.status_code)
            assert rr.status_code == 200
            break
        time.sleep(1)
    else:
        raise RuntimeError("photo never completed")

    r = requests.delete(f"{BASE_URL}/albums/{album_id}/photos/{photo_id}")
    print("delete photo:", r.status_code)
    assert r.status_code in (200, 204)

    r = requests.get(f"{BASE_URL}/albums/{album_id}/photos/{photo_id}")
    print("get deleted photo:", r.status_code, r.text)
    assert r.status_code == 404

    rr = requests.get(url)
    print("deleted file url:", rr.status_code)
    assert rr.status_code != 200

if __name__ == "__main__":
    main()