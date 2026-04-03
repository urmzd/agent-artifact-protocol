Create Celery task definitions for a web application's background job processing.

Include:
- Celery app configuration with Redis broker and result backend
- Email tasks: send_welcome_email, send_password_reset, send_notification_digest (with retry logic)
- Report tasks: generate_daily_report, generate_monthly_summary, export_user_data (with progress tracking)
- Cleanup tasks: purge_expired_sessions, archive_old_records, cleanup_temp_files (with scheduling via beat)
- Proper error handling, retries with exponential backoff, task chaining examples

Use section IDs: config, email-tasks, report-tasks, cleanup-tasks

Use AAP section markers to delineate each major code block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.