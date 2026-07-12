"""Tests for the extended task-manager features: due dates, tags, search,
and sorting.
"""

import pytest

from taskmanager import core


def test_add_task_stores_due_date():
    tasks = core.add_task([], "First", due_date="2026-08-01")
    assert tasks[0]["due_date"] == "2026-08-01"


def test_add_task_due_date_defaults_to_none():
    tasks = core.add_task([], "First")
    assert tasks[0]["due_date"] is None


def test_add_task_rejects_invalid_due_date():
    with pytest.raises(ValueError):
        core.add_task([], "First", due_date="not-a-date")


def test_add_task_stores_tags_and_strips_blanks():
    tasks = core.add_task([], "First", tags=["work", "  urgent  ", "", "   "])
    assert tasks[0]["tags"] == ["work", "urgent"]


def test_add_task_tags_default_to_empty_list():
    tasks = core.add_task([], "First")
    assert tasks[0]["tags"] == []


def test_search_tasks_matches_title_case_insensitively():
    tasks = core.add_task([], "Write the report")
    tasks = core.add_task(tasks, "Email the team")
    results = core.search_tasks(tasks, "REPORT")
    assert [task["title"] for task in results] == ["Write the report"]


def test_search_tasks_matches_tags():
    tasks = core.add_task([], "First", tags=["work"])
    tasks = core.add_task(tasks, "Second", tags=["home"])
    results = core.search_tasks(tasks, "work")
    assert [task["title"] for task in results] == ["First"]


def test_search_tasks_does_not_mutate_input():
    tasks = core.add_task([], "First")
    original = list(tasks)
    core.search_tasks(tasks, "first")
    assert tasks == original


def test_sort_tasks_by_priority_orders_high_medium_low():
    tasks = core.add_task([], "Low task", priority="low")
    tasks = core.add_task(tasks, "High task", priority="high")
    tasks = core.add_task(tasks, "Medium task", priority="medium")
    sorted_tasks = core.sort_tasks(tasks, by="priority")
    assert [task["title"] for task in sorted_tasks] == [
        "High task",
        "Medium task",
        "Low task",
    ]


def test_sort_tasks_by_due_date_puts_none_last():
    tasks = core.add_task([], "No date")
    tasks = core.add_task(tasks, "Later", due_date="2026-09-01")
    tasks = core.add_task(tasks, "Sooner", due_date="2026-08-01")
    sorted_tasks = core.sort_tasks(tasks, by="due_date")
    assert [task["title"] for task in sorted_tasks] == ["Sooner", "Later", "No date"]


def test_sort_tasks_by_id_is_default():
    tasks = core.add_task([], "First")
    tasks = core.add_task(tasks, "Second")
    sorted_tasks = core.sort_tasks(list(reversed(tasks)))
    assert [task["id"] for task in sorted_tasks] == [1, 2]


def test_sort_tasks_rejects_invalid_key():
    with pytest.raises(ValueError):
        core.sort_tasks([], by="not_a_field")


def test_sort_tasks_does_not_mutate_input():
    tasks = core.add_task([], "First")
    original = list(tasks)
    core.sort_tasks(tasks, by="id")
    assert tasks == original
