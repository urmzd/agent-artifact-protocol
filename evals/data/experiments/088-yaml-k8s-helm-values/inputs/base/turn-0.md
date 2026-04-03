Create a Helm values.yaml for deploying a web application stack to Kubernetes.

Include:
- Global settings: namespace, image registry, environment, labels
- App: image, replicas, resources, env vars, probes, ingress, service account, PDB
- Database: PostgreSQL subchart values (auth, primary, readReplicas, persistence, metrics)
- Monitoring: Prometheus ServiceMonitor, Grafana dashboards, alerting rules
- Redis: subchart values (auth, architecture, persistence)
- Detailed comments explaining each setting

Use section IDs: global, app, database, monitoring

Use AAP section markers to delineate each major block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.