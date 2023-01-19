import os
import re
import argparse
import time
import json
from typing import List, Dict, Any

import numpy as np

from embed import CHARS_PER_TOKEN, EMBEDDING_ENGINE, get_embeddings


EMBEDDING_TOKENS_WINDOW = 512
EMBEDDING_CHARS_WINDOW = round(EMBEDDING_TOKENS_WINDOW * CHARS_PER_TOKEN)
EMBEDDABLE_EXTENSIONS = set(
    [
        "go",
        "ts",
        "tsx",
        "js",
        "jsx",
        "md",
        "markdown",
        "html",
        "graphql",
        "bazel",
        "java",
        "py",
        "rb",
        "php",
    ]
)
EMBEDDABLE_EXTENSIONLESS_FILES = set(["dockerfile", "license"])
EXCLUDED_PATHS = ["/__fixtures__/", "/testdata/", "/mocks"]
MAX_FILE_SIZE_BYTES = 1000000  # 1MB
FILESYSTEM_SAFE_NAME_REGEXP = re.compile(r"[^0-9a-zA-Z]")


def get_filesystem_safe_codebase_id(codebase_id: str):
    return FILESYSTEM_SAFE_NAME_REGEXP.sub("_", codebase_id)


def chunk_text(text: str, chunk_size: int) -> List[Dict[str, Any]]:
    chunks = []
    for start in range(0, len(text), chunk_size):
        end = min(len(text), start + chunk_size)
        chunks.append({"start": start, "end": end, "text": text[start:end]})
    return chunks


def is_file_in_excluded_path(file_path: str) -> bool:
    for excluded_directory in EXCLUDED_PATHS:
        if excluded_directory in file_path.lower():
            return True
    return False


def stream_file_chunks(codebase_path: str):
    for root, _, files in os.walk(codebase_path):
        for name in files:
            _, ext = os.path.splitext(name)
            # Remove dot and convert to lowercase
            ext_lower = ext[1:].lower()
            if (
                ext_lower in EMBEDDABLE_EXTENSIONS
                or name.lower() in EMBEDDABLE_EXTENSIONLESS_FILES
            ):
                file_path = os.path.join(root, name)
                size = os.path.getsize(file_path)

                if size >= MAX_FILE_SIZE_BYTES:
                    continue

                if is_file_in_excluded_path(file_path):
                    continue

                with open(file_path, encoding="utf-8") as f:
                    file_contents = f.read()

                file_chunks = chunk_text(file_contents, EMBEDDING_CHARS_WINDOW)
                for chunk in file_chunks:
                    relative_path = file_path[len(codebase_path) + 1 :]
                    yield {**chunk, "filePath": relative_path}


def batch_iterator(iterator, batch_size: int) -> List:
    batch = []
    for item in iterator:
        batch.append(item)

        if len(batch) == batch_size:
            yield batch
            batch = []

    if len(batch) > 0:
        yield batch


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--codebase-id", dest="codebase_id")
    parser.add_argument("--codebase-path", dest="codebase_path")
    parser.add_argument("--output-dir", dest="output_dir")
    args = parser.parse_args()

    embeddings_metadata, embeddings = [], []

    t_start = time.time()
    for batch in batch_iterator(stream_file_chunks(args.codebase_path), 512):
        t_batch_start = time.time()
        batch_embeddings = get_embeddings(
            [chunk["text"] for chunk in batch], engine=EMBEDDING_ENGINE
        )
        print("Batch embedding time:", time.time() - t_batch_start)
        embeddings_metadata.extend(batch)
        embeddings.extend(batch_embeddings)

    print("Total embedding time:", time.time() - t_start)

    fs_safe_codebase_id = get_filesystem_safe_codebase_id(args.codebase_id)
    embeddings_metadata_path = os.path.join(
        args.output_dir, f"{fs_safe_codebase_id}_embeddings_metadata.json"
    )
    with open(embeddings_metadata_path, "w", encoding="utf-8") as f:
        json.dump(embeddings_metadata, f)

    embeddings_path = os.path.join(
        args.output_dir, f"{fs_safe_codebase_id}_embeddings.npy"
    )
    np.save(embeddings_path, np.array(embeddings))
