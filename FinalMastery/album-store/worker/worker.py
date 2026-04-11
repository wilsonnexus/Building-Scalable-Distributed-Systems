import os
import shutil
import sqlite3
import time

DB_PATH = os.getenv("DB_PATH", "/app/data/app.db")
PUBLIC_BASE_URL = os.getenv("PUBLIC_BASE_URL", "http://localhost")
POLL_INTERVAL = float(os.getenv("POLL_INTERVAL", "0.5"))

def get_conn():
    conn = sqlite3.connect(DB_PATH, timeout=30, isolation_level=None)
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA journal_mode=WAL;")
    return conn

def claim_job(conn):
    conn.execute("BEGIN IMMEDIATE")
    row = conn.execute(
        "SELECT job_id, photo_id, album_id, upload_path, public_path FROM jobs WHERE status = 'queued' ORDER BY job_id LIMIT 1"
    ).fetchone()

    if row is None:
        conn.execute("COMMIT")
        return None

    conn.execute(
        "UPDATE jobs SET status = 'processing', attempts = attempts + 1 WHERE job_id = ? AND status = 'queued'",
        (row["job_id"],),
    )
    conn.execute("COMMIT")
    return row

def complete_job(conn, row):
    photo_id = row["photo_id"]
    upload_path = row["upload_path"]
    public_path = row["public_path"]

    os.makedirs(os.path.dirname(public_path), exist_ok=True)

    time.sleep(1.0)

    shutil.copyfile(upload_path, public_path)
    url = f"{PUBLIC_BASE_URL.rstrip('/')}/media/{photo_id}"

    conn.execute(
        "UPDATE photos SET status = 'completed', url = ?, public_path = ? WHERE photo_id = ?",
        (url, public_path, photo_id),
    )
    conn.execute(
        "UPDATE jobs SET status = 'completed', error = '' WHERE photo_id = ?",
        (photo_id,),
    )

def fail_job(conn, row, err):
    conn.execute(
        "UPDATE photos SET status = 'failed' WHERE photo_id = ?",
        (row["photo_id"],),
    )
    conn.execute(
        "UPDATE jobs SET status = 'failed', error = ? WHERE photo_id = ?",
        (str(err), row["photo_id"]),
    )

def main():
    print("worker starting")
    while True:
        conn = get_conn()
        try:
            row = claim_job(conn)
            if row is None:
                conn.close()
                time.sleep(POLL_INTERVAL)
                continue
            try:
                complete_job(conn, row)
            except Exception as e:
                fail_job(conn, row, e)
        except Exception as e:
            print("worker loop error:", e)
        finally:
            conn.close()
        time.sleep(POLL_INTERVAL)

if __name__ == "__main__":
    main()