Create a self-contained HTML Kanban board for project management.

Include:
- Board header with project name "Sprint 24", member avatars, and filter buttons
- 4 columns: Backlog (6 cards), In Progress (4 cards), In Review (3 cards), Done (5 cards)
- Each card has: title, assignee, priority tag (low/medium/high/critical), due date, labels
- Cards should have realistic software development task names
- Basic drag hint styling (grab cursor)

Use section IDs: board-header, col-backlog, col-in-progress, col-review, col-done

Use AAP section markers to delineate each major content block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.

All CSS inline. Horizontal scrolling layout with colored column headers.

Output raw code only. No markdown fences, no explanation.