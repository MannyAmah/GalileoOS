"""Hello Agent entry point.

Stage 0 skeleton. Week 3 wires the real flow.
"""

from __future__ import annotations


def name() -> str:
    """Return the agent identifier used by the runner."""
    return "galileo-hello-agent"


if __name__ == "__main__":
    print(f"{name()} stage0 skeleton")
