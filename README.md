# Kube-Snooze

A Kubernetes operator that automatically scales down or deletes resources during quiet times to save costs and resources.

## 🔧 Core Features

### Custom Resource Definition (CRD)
- **SnoozeWindow CRD** to define snooze rules (resources, schedules, conditions)
- Example fields: resources, namespace, labelSelector, snoozeSchedule, wakeSchedule, timezone

### Schedule-based Snoozing
- **Cron expression support** for flexible scheduling
- **Timezone support** for global deployments
- **Weekday/weekend differentiation**
- **RFC3339 time format** support for one-time events

### Resource Types Handling
- Support for **Deployments**, **StatefulSets**, **CronJobs**, **Jobs**, **Pods**
- Optionally **scale to zero**, **delete**, or **patch** resources
- **Backup & restore** original replicas/state

### Wake Mechanism
- **Restore replicas** to previous state from annotations
- **Wake up at scheduled time** or on-demand
- **State preservation** during snooze periods

### Annotation & Label Opt-in/Opt-out
- Use annotations (e.g., `kube-snooze/enabled: true`) to select resources for snoozing
- **Label selectors** to include/exclude resources
- **Policy-based management** with multiple snooze windows

## 🚀 Quick Start

### 1. Install the Operator

```bash
# Apply CRDs
kubectl apply -f config/crd/bases/

# Apply RBAC
kubectl apply -f config/rbac/

# Deploy the operator
kubectl apply -f config/manager/
```

### 2. Create a SnoozeWindow

```yaml
apiVersion: scheduling.codeacme.org/v1alpha1
kind: SnoozeWindow
metadata:
  name: weekend-snooze
  namespace: default
spec:
  labelSelector:
    app: "my-app"
    environment: "development"
  
  snoozeSchedule:
    cronExpression: "0 18 * * 1-5"  # Weekdays at 6 PM
  
  wakeSchedule:
    cronExpression: "0 8 * * 1-5"   # Weekdays at 8 AM
  
  timezone: "America/New_York"
  
  resourceTypes:
    - kind: "Deployment"
      apiVersion: "apps/v1"
      scaleToZero: true
  
  snoozeAction:
    scaleToZero: true
  
  backupConfig:
    storeInAnnotations: true
```

### 3. Annotate Resources

Add the following annotation to resources you want to manage:

```yaml
metadata:
  annotations:
    kube-snooze/enabled: "true"
```

## 📋 Usage Examples

### Basic Weekend Snoozing

```yaml
apiVersion: scheduling.codeacme.org/v1alpha1
kind: SnoozeWindow
metadata:
  name: weekend-snooze
spec:
  labelSelector:
    environment: "development"
  
  snoozeSchedule:
    cronExpression: "0 18 * * 5"  # Friday at 6 PM
  
  wakeSchedule:
    cronExpression: "0 8 * * 1"   # Monday at 8 AM
  
  resourceTypes:
    - kind: "Deployment"
      apiVersion: "apps/v1"
      scaleToZero: true
  
  snoozeAction:
    scaleToZero: true
```

### Night-time Snoozing with Custom Patch

```yaml
apiVersion: scheduling.codeacme.org/v1alpha1
kind: SnoozeWindow
metadata:
  name: night-snooze
spec:
  labelSelector:
    app: "database"
  
  snoozeSchedule:
    cronExpression: "0 22 * * *"  # Daily at 10 PM
  
  wakeSchedule:
    cronExpression: "0 6 * * *"   # Daily at 6 AM
  
  resourceTypes:
    - kind: "StatefulSet"
      apiVersion: "apps/v1"
      scaleToZero: true
  
  snoozeAction:
    patch:
      type: "strategic"
      data: '{"spec":{"replicas":0,"template":{"spec":{"containers":[{"name":"db","resources":{"requests":{"cpu":"10m","memory":"10Mi"}}}]}}}}'
```

### Delete Jobs and CronJobs

```yaml
apiVersion: scheduling.codeacme.org/v1alpha1
kind: SnoozeWindow
metadata:
  name: cleanup-jobs
spec:
  labelSelector:
    app: "batch-processing"
  
  snoozeSchedule:
    cronExpression: "0 0 * * 0"  # Sunday at midnight
  
  wakeSchedule:
    cronExpression: "0 8 * * 1"  # Monday at 8 AM
  
  resourceTypes:
    - kind: "CronJob"
      apiVersion: "batch/v1"
      delete: true
    - kind: "Job"
      apiVersion: "batch/v1"
      delete: true
  
  snoozeAction:
    delete: true
```

## 🔍 Monitoring

### Check SnoozeWindow Status

```bash
kubectl get snoozewindows
kubectl describe snoozewindow weekend-snooze
```

### View Managed Resources

```bash
# List resources with snooze annotations
kubectl get deployments,statefulsets,cronjobs -A -o jsonpath='{range .items[?(@.metadata.annotations.kube-snooze/enabled=="true")]}{.kind}/{.metadata.namespace}/{.metadata.name}{"\n"}{end}'
```

### Check Backup State

```bash
# View backup annotations
kubectl get deployment my-app -o jsonpath='{.metadata.annotations.kube-snooze/backup-replicas}'
```

## 🏗️ Architecture

### Controller Logic

1. **Schedule Evaluation**: Uses cron expressions to determine when to snooze/wake
2. **Resource Discovery**: Finds resources matching label selectors and annotations
3. **State Backup**: Stores original state in annotations before snoozing
4. **Action Application**: Scales, deletes, or patches resources as configured
5. **State Restoration**: Restores original state when waking

### Resource Management

- **Deployments/StatefulSets**: Scale to zero replicas
- **CronJobs/Jobs**: Delete entirely
- **Custom Patches**: Apply strategic merge patches
- **State Backup**: Store in annotations or ConfigMaps

### Scheduling

- **Cron Expressions**: Standard cron format support
- **Timezone Support**: Local timezone calculations
- **Weekday/Weekend**: Flexible day-of-week filtering
- **RFC3339**: One-time event scheduling

## 🔧 Configuration

### SnoozeWindow Spec Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | string | Target namespace (optional) |
| `labelSelector` | map[string]string | Resource selector |
| `snoozeSchedule` | ScheduleConfig | When to snooze |
| `wakeSchedule` | ScheduleConfig | When to wake |
| `timezone` | string | Timezone for schedules |
| `resourceTypes` | []ResourceType | Types to manage |
| `snoozeAction` | SnoozeAction | What action to take |
| `backupConfig` | BackupConfig | How to backup state |

### Annotations

| Annotation | Value | Description |
|------------|-------|-------------|
| `kube-snooze/enabled` | "true" | Enable snoozing for this resource |
| `kube-snooze/policy` | "policy-name" | Associate with specific policy |
| `kube-snooze/backup-full-state` | "true" | Backup complete resource state |

## 🚨 Best Practices

1. **Test in Development**: Always test snooze policies in development first
2. **Use Labels**: Use meaningful labels for resource selection
3. **Backup State**: Enable state backup for critical resources
4. **Monitor Schedules**: Verify cron expressions work as expected
5. **Gradual Rollout**: Start with non-critical resources

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0.

