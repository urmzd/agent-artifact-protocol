Create a Go gRPC service implementation for a notification service.

Include:
- Types: Notification, NotificationChannel (email, SMS, push), Template, DeliveryStatus
- Service struct implementing NotificationServiceServer with dependency injection
- Handlers: SendNotification, GetNotificationStatus, ListNotifications, UpdatePreferences, RegisterDevice
- Interceptors: logging, auth, rate limiting, panic recovery
- Helper functions for template rendering and channel-specific delivery

Use section IDs: types, service, handlers, interceptors

Use AAP section markers to delineate each major code block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.