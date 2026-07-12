"""Tests for the task priority feature.

These cover adding a task with a priority (including the default), filtering
tasks by priority, and rejecting invalid priority values.
"""

import pytest

from taskmanager import core


def test_add_task_defaults_to_medium_priority():
    tasks = core.add_task([], "First")
    assert tasks[0]["priority"] == "medium"


def test_add_task_accepts_given_priority():
    tasks = core.add_task([], "First", priority="high")
    assert tasks[0]["priority"] == "high"


def test_add_task_rejects_invalid_priority():
    with pytest.raises(ValueError):
        core.add_task([], "First", priority="urgent")


def test_tasks_with_priority_filters_and_preserves_order():
    tasks = core.add_task([], "First", priority="high")
    tasks = core.add_task(tasks, "Second", priority="low")
    tasks = core.add_task(tasks, "Third", priority="high")

    high_priority = core.tasks_with_priority(tasks, "high")

    assert [task["title"] for task in high_priority] == ["First", "Third"]


def test_tasks_with_priority_does_not_mutate_input():
    tasks = core.add_task([], "First", priority="high")
    original = list(tasks)

    core.tasks_with_priority(tasks, "high")

    assert tasks == original
