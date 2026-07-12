"""Core task operations for the CSO7024 task manager.

Tasks are represented as plain dictionaries so the data is easy to inspect
and to serialise to JSON. A task currently has three fields:

    id     an integer, unique within the list
    title  a non-empty string
    done   a boolean, False when the task is created

Every operation returns a *new* list rather than modifying its argument. This
keeps the functions easy to test and reason about.

In the mid-module assessment you will extend this module with a task
"priority" and the ability to filter by it. The exact specification is in the
README. Do not change the behaviour the existing tests rely on.
"""

from __future__ import annotations

import json
from datetime import date
from pathlib import Path

VALID_PRIORITIES = ("high", "medium", "low")
_PRIORITY_ORDER = {"high": 0, "medium": 1, "low": 2}


def add_task(
    tasks: list[dict],
    title: str,
    priority: str = "medium",
    due_date: str | None = None,
    tags: list[str] | None = None,
) -> list[dict]:
    """Return a new task list with one task appended.

    The new task is given the next integer id (one more than the current
    highest, or 1 for the first task), the supplied title, the given
    ``priority`` (defaulting to ``"medium"``), an optional ``due_date``, an
    optional list of ``tags``, and a ``done`` flag of ``False``.

    Args:
        due_date: an ISO ``YYYY-MM-DD`` date string, or ``None`` if the task
            has no due date.
        tags: a list of tag strings, or ``None`` for no tags. Empty or
            whitespace-only tags are dropped.

    Raises:
        ValueError: if ``title`` is empty or only whitespace, if
            ``priority`` is not one of ``"high"``, ``"medium"`` or ``"low"``,
            or if ``due_date`` is not a valid ``YYYY-MM-DD`` date.
    """
    if title is None or not title.strip():
        raise ValueError("Task title must not be empty")
    if priority not in VALID_PRIORITIES:
        raise ValueError(
            f"Invalid priority {priority!r}; must be one of {VALID_PRIORITIES}"
        )
    if due_date is not None:
        try:
            date.fromisoformat(due_date)
        except ValueError as exc:
            raise ValueError(
                f"Invalid due_date {due_date!r}; must be an ISO date (YYYY-MM-DD)"
            ) from exc
    clean_tags = [tag.strip() for tag in (tags or []) if tag and tag.strip()]
    next_id = max((task["id"] for task in tasks), default=0) + 1
    new_task = {
        "id": next_id,
        "title": title.strip(),
        "done": False,
        "priority": priority,
        "due_date": due_date,
        "tags": clean_tags,
    }
    return tasks + [new_task]


def tasks_with_priority(tasks: list[dict], priority: str) -> list[dict]:
    """Return a new list of the tasks whose priority equals ``priority``.

    The input list is not modified, and the original order is preserved.
    """
    return [task for task in tasks if task["priority"] == priority]


def search_tasks(tasks: list[dict], query: str) -> list[dict]:
    """Return a new list of the tasks whose title or tags match ``query``.

    The match is a case-insensitive substring match against the task title
    and against each of its tags. The input list is not modified, and the
    original order is preserved.
    """
    needle = query.strip().lower()
    matches = []
    for task in tasks:
        if needle in task["title"].lower():
            matches.append(task)
            continue
        if any(needle in tag.lower() for tag in task.get("tags", [])):
            matches.append(task)
    return matches


def sort_tasks(tasks: list[dict], by: str = "id") -> list[dict]:
    """Return a new list of ``tasks`` sorted by ``by``.

    Supported values for ``by``:
        "id"        ascending task id (the default).
        "priority"  "high" first, then "medium", then "low".
        "due_date"  ascending due date, with tasks that have no due date
                    sorted last.

    The input list is not modified.

    Raises:
        ValueError: if ``by`` is not one of the supported values.
    """
    if by == "id":
        key = lambda task: task["id"]
    elif by == "priority":
        key = lambda task: _PRIORITY_ORDER[task["priority"]]
    elif by == "due_date":
        key = lambda task: (task["due_date"] is None, task["due_date"])
    else:
        raise ValueError(f"Invalid sort key {by!r}; must be one of 'id', 'priority', 'due_date'")
    return sorted(tasks, key=key)


def complete_task(tasks: list[dict], task_id: int) -> list[dict]:
    """Return a new task list with the task of ``task_id`` marked done.

    Raises:
        KeyError: if no task has the given id.
    """
    if not any(task["id"] == task_id for task in tasks):
        raise KeyError(f"No task with id {task_id}")
    return [
        {**task, "done": True} if task["id"] == task_id else task
        for task in tasks
    ]


def remove_task(tasks: list[dict], task_id: int) -> list[dict]:
    """Return a new task list with the task of ``task_id`` removed.

    Raises:
        KeyError: if no task has the given id.
    """
    if not any(task["id"] == task_id for task in tasks):
        raise KeyError(f"No task with id {task_id}")
    return [task for task in tasks if task["id"] != task_id]


def load_tasks(path: Path) -> list[dict]:
    """Load tasks from a JSON file, returning an empty list if it is absent."""
    if not path.exists():
        return []
    return json.loads(path.read_text(encoding="utf-8"))


def save_tasks(tasks: list[dict], path: Path) -> None:
    """Write tasks to a JSON file as indented JSON."""
    path.write_text(json.dumps(tasks, indent=2), encoding="utf-8")
