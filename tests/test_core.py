"""Tests for the existing task operations.

These tests must continue to pass after you add the priority feature, so be
careful not to change the behaviour they rely on. Add your new tests for the
priority feature in a separate test file (for example, test_priority.py).

Run all tests with:
    pytest
"""

import pytest

from taskmanager import core


def test_add_task_appends_with_incrementing_ids():
    tasks = []
    tasks = core.add_task(tasks, "First")
    tasks = core.add_task(tasks, "Second")
    assert [task["id"] for task in tasks] == [1, 2]
    assert tasks[0]["title"] == "First"
    assert tasks[1]["done"] is False


def test_add_task_does_not_mutate_input():
    original: list[dict] = []
    core.add_task(original, "First")
    assert original == []


def test_add_task_strips_whitespace():
    tasks = core.add_task([], "  Buy milk  ")
    assert tasks[0]["title"] == "Buy milk"


def test_add_task_rejects_empty_title():
    with pytest.raises(ValueError):
        core.add_task([], "   ")


def test_complete_task_marks_done():
    tasks = core.add_task([], "First")
    tasks = core.complete_task(tasks, 1)
    assert tasks[0]["done"] is True


def test_complete_task_unknown_id_raises():
    with pytest.raises(KeyError):
        core.complete_task([], 99)


def test_remove_task_removes_by_id():
    tasks = core.add_task([], "First")
    tasks = core.remove_task(tasks, 1)
    assert tasks == []


def test_remove_task_unknown_id_raises():
    with pytest.raises(KeyError):
        core.remove_task([], 99)
