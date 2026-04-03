Create a self-contained HTML service status page.

Include:
- Header with company name and overall status indicator (All Systems Operational)
- Current status banner with last updated timestamp
- Services list (10 services): API, Web App, Database, CDN, Email, Search, Auth, Payments, Storage, Queue — each with status dot and response time
- Recent incidents section (3 incidents with timeline: investigating, identified, monitoring, resolved)
- 90-day uptime history bars for each service (visual representation with colored segments)

Use section IDs: header, current-status, services, incidents, uptime-history

Use AAP section markers to delineate each major content block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.

All CSS inline. Clean, minimal design with green/yellow/red status colors.

Output raw code only. No markdown fences, no explanation.