"""A small command-line interface for the task manager.

Tasks are stored in ``tasks.json`` in the current working directory.

Examples:
    python -m taskmanager.cli add "Write the report"
    python -m taskmanager.cli list
    python -m taskmanager.cli done 1
    python -m taskmanager.cli remove 1
"""

from __future__ import annotations

import argparse
from pathlib import Path

from taskmanager import core

DEFAULT_STORE = Path("tasks.json")


def _format(task: dict) -> str:
    box = "[x]" if task["done"] else "[ ]"
    due = f" due:{task['due_date']}" if task.get("due_date") else ""
    tags = f" #{','.join(task['tags'])}" if task.get("tags") else ""
    return f"{task['id']:>3} {box} {task['title']} ({task['priority']}){due}{tags}"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="A small task manager.")
    sub = parser.add_subparsers(dest="command", required=True)

    add_p = sub.add_parser("add", help="Add a new task")
    add_p.add_argument("title", help="The task title")
    add_p.add_argument(
        "--priority",
        choices=core.VALID_PRIORITIES,
        default="medium",
        help="The task priority (default: medium)",
    )
    add_p.add_argument(
        "--due-date",
        default=None,
        help="Due date as an ISO date, e.g. 2026-07-31",
    )
    add_p.add_argument(
        "--tags",
        default=None,
        help="Comma-separated tags, e.g. work,urgent",
    )

    list_p = sub.add_parser("list", help="List all tasks")
    list_p.add_argument(
        "--priority",
        choices=core.VALID_PRIORITIES,
        default=None,
        help="Only show tasks with this priority",
    )
    list_p.add_argument(
        "--search",
        default=None,
        help="Only show tasks whose title or tags match this text",
    )
    list_p.add_argument(
        "--sort-by",
        choices=("id", "priority", "due_date"),
        default="id",
        help="Sort the listed tasks by this field (default: id)",
    )

    done_p = sub.add_parser("done", help="Mark a task as done")
    done_p.add_argument("task_id", type=int, help="The id of the task")

    remove_p = sub.add_parser("remove", help="Remove a task")
    remove_p.add_argument("task_id", type=int, help="The id of the task")

    args = parser.parse_args(argv)
    tasks = core.load_tasks(DEFAULT_STORE)

    if args.command == "add":
        tags = args.tags.split(",") if args.tags else None
        tasks = core.add_task(
            tasks,
            args.title,
            priority=args.priority,
            due_date=args.due_date,
            tags=tags,
        )
        core.save_tasks(tasks, DEFAULT_STORE)
        print(f"Added: {args.title}")
    elif args.command == "list":
        shown = tasks if args.priority is None else core.tasks_with_priority(tasks, args.priority)
        if args.search:
            shown = core.search_tasks(shown, args.search)
        shown = core.sort_tasks(shown, by=args.sort_by)
        if not shown:
            print("No tasks yet.")
        for task in shown:
            print(_format(task))
    elif args.command == "done":
        tasks = core.complete_task(tasks, args.task_id)
        core.save_tasks(tasks, DEFAULT_STORE)
        print(f"Completed task {args.task_id}")
    elif args.command == "remove":
        tasks = core.remove_task(tasks, args.task_id)
        core.save_tasks(tasks, DEFAULT_STORE)
        print(f"Removed task {args.task_id}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
