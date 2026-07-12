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
    return f"{task['id']:>3} {box} {task['title']}"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="A small task manager.")
    sub = parser.add_subparsers(dest="command", required=True)

    add_p = sub.add_parser("add", help="Add a new task")
    add_p.add_argument("title", help="The task title")

    sub.add_parser("list", help="List all tasks")

    done_p = sub.add_parser("done", help="Mark a task as done")
    done_p.add_argument("task_id", type=int, help="The id of the task")

    remove_p = sub.add_parser("remove", help="Remove a task")
    remove_p.add_argument("task_id", type=int, help="The id of the task")

    args = parser.parse_args(argv)
    tasks = core.load_tasks(DEFAULT_STORE)

    if args.command == "add":
        tasks = core.add_task(tasks, args.title)
        core.save_tasks(tasks, DEFAULT_STORE)
        print(f"Added: {args.title}")
    elif args.command == "list":
        if not tasks:
            print("No tasks yet.")
        for task in tasks:
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
