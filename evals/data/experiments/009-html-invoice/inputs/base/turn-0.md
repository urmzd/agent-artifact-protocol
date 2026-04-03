Create a self-contained HTML invoice document.

Include:
- Header with company name "TechCorp Solutions", logo placeholder, invoice number INV-2024-0847
- From/To addresses side by side
- Line items table with 8 items (description, quantity, rate, amount)
- Totals section (subtotal, discount 10%, tax 8.5%, total due)
- Payment terms and bank details at the bottom

Use section IDs: header, addresses, line-items, totals

Use AAP section markers to delineate each major content block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.

All CSS inline. Print-friendly layout, max width 800px.

Output raw code only. No markdown fences, no explanation.