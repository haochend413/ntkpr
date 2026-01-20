from typing import Any
import httpx
from mcp.server.fastmcp import FastMCP
from sqlmodel import Field, Session, SQLModel, create_engine, select
from pathlib import Path
from models import Note, NoteBase
from langchain_ollama import OllamaEmbeddings
from pathlib import Path
from qdrant_client import QdrantClient
from langchain_qdrant import QdrantVectorStore
from contextlib import asynccontextmanager
from fastapi import FastAPI

# define lifespan;


raw_data = {"notes": []}


# fetch everything;
def fetchData(session: Session):
    global raw_data
    with session:
        raw_data["notes"] = session.exec(select(Note)).all()


# vectorization;
embedding_model = OllamaEmbeddings(model="nomic-embed-text")
qdrant_client = QdrantClient(host="localhost", port=6333)
vectorstore = None  # Will init on ingest


# These will be set after vectorstore is initialized
# llm = None
# retriever = None
# qa_chain = None


# run at startup;


def ingest_notes(session: Session):
    # erase previous record before re-starting;
    qdrant_client.delete_collection(collection_name="notes")
    notes = raw_data["notes"]
    docs = []
    payload = []
    for note in notes:
        text = f"Note: {note.content}"
        docs.append(text)
        payload.append({"note_id": note.id})

    global vectorstore

    vectorstore = QdrantVectorStore.from_texts(
        texts=docs,
        embedding=embedding_model,
        location="http://localhost:6333",
        collection_name="notes",
        metadatas=payload,
    )
    return {"status": "indexed"}


@asynccontextmanager
async def lifespan(app: FastAPI):
    print("fetching embeddings ... ")
    fetchData(Session(engine))
    ingest_notes(Session(engine))
    yield


# Initialize FastMCP server
mcp = FastMCP("notes", lifespan=lifespan)

# Constants
sqlite_file_name = (
    Path(__file__).parent.parent.parent / "ntkpr" / "cmd" / "notes" / "notes.db"
).as_posix()
sqlite_url = f"sqlite:///{sqlite_file_name}"
USER_AGENT = "mantis/1.0"
engine = create_engine(sqlite_url, echo=True)


def get_session():
    with Session(engine) as session:
        yield session


# now, add functions for READ_ONLY first (Later enable agent to change the database;)

# query helper functions;


@mcp.tool(description="Get a list of all notes from the database.")
def read_notes(input: str = ""):
    """Get all notes from the database that haven't been deleted."""
    with Session(engine) as session:
        notes = session.exec(
            select(Note).where(Note.deleted_at.is_(False) | Note.deleted_at.is_(None))
        ).all()
        return [note.model_dump(mode="python") for note in notes]


# @mcp.tool(description="Get a list of all notes with a given ID")
# def read_notes_with_id():
#     pass


@mcp.tool(description="Search notes based on similar meanings and contents. ")
def search_notes(query: str):
    """Get a list of notes based on similar meanings and contenst.

    Args:
        input: query content to be compared as a string.
    """
    results = vectorstore.similarity_search(query, k=10)
    return [doc.page_content for doc in results]


# @mcp.tool(description="Search related notes by semantic meaning.")
# def rag_search(q: str):
#     retriever = vectorstore.as_retriever(search_kwargs={"k": 10})
#     qa_chain = RetrievalQA.from_chain_type(llm=llm, retriever=retriever)


# run server;
if __name__ == "__main__":
    print(sqlite_file_name)
    print("starting...")
    mcp.run(transport="stdio")
    # print("started")
