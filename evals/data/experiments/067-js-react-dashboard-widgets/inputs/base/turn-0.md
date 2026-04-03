Create React dashboard widget components.

Include:
- StatCard: displays metric name, value, trend (up/down with percentage), icon placeholder
- ChartWidget: wrapper with title, date range selector, placeholder for chart area with sample SVG sparkline
- ActivityFeed: scrollable list of 10 activity items (avatar, user, action, timestamp, details)
- QuickActions: grid of 6 action buttons with icon and label (New Project, Add User, Generate Report, etc.)
- All components with PropTypes and default props
- Inline styles using style objects

Use section IDs: stat-card, chart-widget, activity-feed, quick-actions

Use AAP section markers to delineate each major code block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.