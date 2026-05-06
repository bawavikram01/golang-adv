// =============================================================================
// LESSON 11: DEPLOYMENT PATTERNS — Containers, Kubernetes, Strategies
// =============================================================================
//
// Microservices without proper deployment = chaos.
// You're deploying 10-50 services independently. How?
//
// THIS LESSON COVERS:
//   1. Containerization (Docker)
//   2. Container Orchestration (Kubernetes)
//   3. Deployment Strategies (Blue-Green, Canary, Rolling)
//   4. Service Mesh (Istio, Linkerd)
//   5. GitOps, CI/CD, Infrastructure as Code
// =============================================================================

package main

import (
	"fmt"
	"time"
)

// =============================================================================
// CONCEPT 1: Docker — Containerization
// =============================================================================
//
// Container = lightweight, portable, self-contained package of your service.
// Contains: binary + dependencies + config. Nothing more.
//
// WHY CONTAINERS:
//   ✅ "Works on my machine" → works EVERYWHERE
//   ✅ Consistent environments (dev = staging = prod)
//   ✅ Fast startup (seconds, not minutes like VMs)
//   ✅ Resource isolation (CPU, memory limits)
//   ✅ Immutable deployments (same image everywhere)
//
// BEST PRACTICES FOR GO DOCKER IMAGES:
//   1. Multi-stage build (build in one stage, copy binary to scratch/alpine)
//   2. Use scratch or distroless (no shell, no package manager = smaller attack surface)
//   3. Run as non-root user
//   4. Pin dependency versions
//   5. Use .dockerignore

type Dockerfile struct {
	Stages []DockerStage
}

type DockerStage struct {
	Name         string
	BaseImage    string
	Instructions []string
}

func OptimalGoDockerfile() Dockerfile {
	return Dockerfile{
		Stages: []DockerStage{
			{
				Name:      "builder",
				BaseImage: "golang:1.22-alpine",
				Instructions: []string{
					"WORKDIR /app",
					"COPY go.mod go.sum ./",
					"RUN go mod download",
					"COPY . .",
					"RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o /app/server ./cmd/server",
				},
			},
			{
				Name:      "final",
				BaseImage: "gcr.io/distroless/static-debian12:nonroot",
				Instructions: []string{
					"COPY --from=builder /app/server /server",
					"COPY --from=builder /app/configs /configs",
					"EXPOSE 8080",
					"USER nonroot:nonroot",
					"ENTRYPOINT [\"/server\"]",
				},
			},
		},
	}
}

// =============================================================================
// CONCEPT 2: Kubernetes — Container Orchestration
// =============================================================================
//
// Kubernetes (K8s) manages containers at scale.
//
// KEY OBJECTS:
//   Pod:        Smallest unit. 1+ containers. Ephemeral.
//   Deployment: Manages pods. Handles scaling, updates, rollbacks.
//   Service:    Stable network address for a set of pods. Load balances.
//   Ingress:    HTTP routing (path/host-based). Edge of the cluster.
//   ConfigMap:  Configuration (non-secret key-value pairs).
//   Secret:     Sensitive config (passwords, tokens). Base64 encoded.
//   HPA:        Horizontal Pod Autoscaler. Scales pods based on CPU/custom metrics.
//   PDB:        Pod Disruption Budget. Ensures minimum pods during maintenance.
//   NetworkPolicy: Firewall rules between pods.
//
// K8S ARCHITECTURE:
//   Control Plane: API Server, etcd, Scheduler, Controller Manager
//   Worker Nodes:  kubelet, kube-proxy, Container Runtime (containerd)

type K8sManifest struct {
	Kind      string
	Name      string
	Namespace string
	Spec      map[string]interface{}
}

func OrderServiceK8sManifests() []K8sManifest {
	return []K8sManifest{
		{
			Kind: "Deployment", Name: "order-service", Namespace: "production",
			Spec: map[string]interface{}{
				"replicas":        3,
				"image":           "myregistry.com/order-service:v1.2.3",
				"cpu_request":     "100m",
				"cpu_limit":       "500m",
				"memory_request":  "128Mi",
				"memory_limit":    "512Mi",
				"liveness_probe":  "/health/live",
				"readiness_probe": "/health/ready",
				"strategy":        "RollingUpdate",
				"max_surge":       "25%",
				"max_unavailable": "25%",
			},
		},
		{
			Kind: "Service", Name: "order-service", Namespace: "production",
			Spec: map[string]interface{}{
				"type":        "ClusterIP",
				"port":        8080,
				"target_port": 8080,
				"selector":    "app=order-service",
			},
		},
		{
			Kind: "HPA", Name: "order-service-hpa", Namespace: "production",
			Spec: map[string]interface{}{
				"min_replicas":       2,
				"max_replicas":       20,
				"cpu_target_percent": 70,
			},
		},
		{
			Kind: "Ingress", Name: "order-service-ingress", Namespace: "production",
			Spec: map[string]interface{}{
				"host":    "api.example.com",
				"path":    "/api/v1/orders",
				"backend": "order-service:8080",
				"tls":     true,
			},
		},
	}
}

// =============================================================================
// CONCEPT 3: Deployment Strategies
// =============================================================================
//
// How do you update a running service without downtime?
//
// ┌────────────────────┬───────────────────────────────────────────────────┐
// │ Strategy           │ Description                                       │
// ├────────────────────┼───────────────────────────────────────────────────┤
// │ Rolling Update     │ Replace pods one by one. Default in K8s.          │
// │                    │ Old and new versions run simultaneously briefly.  │
// │                    │ Simple, but can't easily roll back.               │
// ├────────────────────┼───────────────────────────────────────────────────┤
// │ Blue-Green         │ Two identical environments: Blue (current),       │
// │                    │ Green (new). Switch all traffic instantly.        │
// │                    │ Easy rollback (switch back). 2x resources.        │
// ├────────────────────┼───────────────────────────────────────────────────┤
// │ Canary             │ Route 5% traffic to new version, monitor.         │
// │                    │ If healthy, gradually increase to 100%.           │
// │                    │ Best for risk reduction. Needs good observability.│
// ├────────────────────┼───────────────────────────────────────────────────┤
// │ A/B Testing        │ Route by user segment (not random).              │
// │                    │ "All premium users get new version."             │
// │                    │ Needs feature flags or smart routing.            │
// ├────────────────────┼───────────────────────────────────────────────────┤
// │ Shadow/Dark Launch │ Send copy of production traffic to new version.  │
// │                    │ New version processes but responses are discarded.│
// │                    │ Test performance/correctness with real traffic.   │
// └────────────────────┴───────────────────────────────────────────────────┘

type DeploymentStrategy struct {
	Name         string
	Description  string
	RollbackTime string
	ResourceCost string
	Risk         string
	BestFor      string
}

var strategies = []DeploymentStrategy{
	{
		Name:         "Rolling Update",
		Description:  "Replace pods gradually (old → new)",
		RollbackTime: "Minutes (new rolling update)",
		ResourceCost: "1x (same pods, swapped gradually)",
		Risk:         "Medium (both versions serve traffic briefly)",
		BestFor:      "Simple services, backward-compatible changes",
	},
	{
		Name:         "Blue-Green",
		Description:  "Two full environments, instant switch",
		RollbackTime: "Seconds (switch back to blue)",
		ResourceCost: "2x (both environments running)",
		Risk:         "Low (instant rollback)",
		BestFor:      "Critical services, database migrations",
	},
	{
		Name:         "Canary",
		Description:  "Gradual traffic shift: 5% → 25% → 50% → 100%",
		RollbackTime: "Seconds (shift back to 0%)",
		ResourceCost: "1.05-1.5x (few extra pods)",
		Risk:         "Very Low (only small % of traffic affected)",
		BestFor:      "High-traffic services, new features",
	},
	{
		Name:         "Shadow/Dark Launch",
		Description:  "Mirror traffic, discard new version's responses",
		RollbackTime: "N/A (doesn't serve real users)",
		ResourceCost: "2x (mirrored traffic)",
		Risk:         "None (responses discarded)",
		BestFor:      "Performance testing, machine learning models",
	},
}

// --- Canary Deployment Simulation ---

type CanaryDeployment struct {
	ServiceName string
	OldVersion  string
	NewVersion  string
	Steps       []CanaryStep
	CurrentStep int
}

type CanaryStep struct {
	TrafficPercent int
	Duration       time.Duration
	MetricCheck    string // what to verify before proceeding
}

func NewCanaryDeployment(service, oldVer, newVer string) *CanaryDeployment {
	return &CanaryDeployment{
		ServiceName: service,
		OldVersion:  oldVer,
		NewVersion:  newVer,
		Steps: []CanaryStep{
			{TrafficPercent: 5, Duration: 5 * time.Minute, MetricCheck: "error_rate < 1%"},
			{TrafficPercent: 25, Duration: 10 * time.Minute, MetricCheck: "error_rate < 1% && p99 < 500ms"},
			{TrafficPercent: 50, Duration: 15 * time.Minute, MetricCheck: "error_rate < 0.5% && p99 < 500ms"},
			{TrafficPercent: 100, Duration: 0, MetricCheck: "fully promoted"},
		},
	}
}

func (c *CanaryDeployment) Simulate() {
	fmt.Printf("  Canary: %s %s → %s\n", c.ServiceName, c.OldVersion, c.NewVersion)
	for i, step := range c.Steps {
		fmt.Printf("  Step %d: %d%% traffic to %s (check: %s, wait: %v)\n",
			i+1, step.TrafficPercent, c.NewVersion, step.MetricCheck, step.Duration)
	}
}

// =============================================================================
// CONCEPT 4: Service Mesh
// =============================================================================
//
// A SERVICE MESH is an infrastructure layer that handles service-to-service
// communication. It's a network of proxies (sidecars) alongside each service.
//
// SIDECAR PROXY: A small proxy (Envoy) deployed alongside each service pod.
//   All traffic in/out goes through the sidecar.
//   The service doesn't know the sidecar exists.
//
// WHAT IT PROVIDES:
//   ✅ mTLS (automatic encryption between all services)
//   ✅ Traffic management (canary, circuit breaker, retry, timeout)
//   ✅ Observability (distributed tracing, metrics — without code changes)
//   ✅ Access control (which service can call which)
//   ✅ Rate limiting per service
//
// ARCHITECTURE:
//   Data Plane:  Sidecar proxies (Envoy) handle actual traffic
//   Control Plane: Configures proxies (Istiod for Istio, linkerd2-proxy for Linkerd)
//
// POPULAR MESHES:
//   Istio:   Feature-rich, complex, Envoy-based. Most popular.
//   Linkerd: Simple, lightweight, Rust-based proxy. Easy to start with.
//   Consul Connect: HashiCorp's mesh. Works outside K8s too.

type ServiceMeshConfig struct {
	Mesh        string
	Features    []string
	ProxyType   string
	ProxyCPU    string
	ProxyMemory string
}

var meshComparison = []ServiceMeshConfig{
	{
		Mesh: "Istio", ProxyType: "Envoy", ProxyCPU: "~100m", ProxyMemory: "~128Mi",
		Features: []string{"mTLS", "traffic management", "observability", "policy", "multi-cluster"},
	},
	{
		Mesh: "Linkerd", ProxyType: "linkerd2-proxy (Rust)", ProxyCPU: "~10m", ProxyMemory: "~20Mi",
		Features: []string{"mTLS", "traffic splitting", "observability", "retries"},
	},
	{
		Mesh: "Consul Connect", ProxyType: "Envoy", ProxyCPU: "~50m", ProxyMemory: "~64Mi",
		Features: []string{"mTLS", "intentions", "multi-platform", "service discovery"},
	},
}

// =============================================================================
// CONCEPT 5: GitOps
// =============================================================================
//
// GitOps = Git is the single source of truth for infrastructure and deployments.
//
// FLOW:
//   1. Developer pushes code → CI runs tests → builds Docker image → pushes to registry
//   2. Developer updates K8s manifests in Git (image tag, config)
//   3. GitOps operator (ArgoCD/Flux) detects Git change
//   4. Operator syncs cluster state to match Git
//   5. If something breaks: git revert → cluster auto-reverts
//
// TOOLS: ArgoCD, FluxCD, Kustomize, Helm
//
// PRINCIPLE: "If it's not in Git, it doesn't exist."

type GitOpsFlow struct {
	Step        int
	Action      string
	Tool        string
	Description string
}

var gitOpsSteps = []GitOpsFlow{
	{1, "Push code", "GitHub", "Developer merges PR to main branch"},
	{2, "CI pipeline", "GitHub Actions", "Lint, test, build Docker image, push to registry"},
	{3, "Update manifest", "Kustomize/Helm", "Update image tag in k8s manifests"},
	{4, "Detect change", "ArgoCD", "Watches Git repo, detects new manifest"},
	{5, "Sync cluster", "ArgoCD", "Applies new manifests to Kubernetes cluster"},
	{6, "Verify health", "ArgoCD", "Checks deployment health, rollbacks if unhealthy"},
}

// =============================================================================
// CONCEPT 6: CI/CD Pipeline for Microservices
// =============================================================================
//
// CHALLENGE: With 20 services, you need 20 CI/CD pipelines.
//            Or: a monorepo with smart change detection.
//
// MONO-REPO: All services in one Git repo.
//   PROS: Atomic changes across services, shared libraries, one PR
//   CONS: CI must detect which services changed, longer builds
//   TOOLS: Bazel, Nx, Turborepo
//
// MULTI-REPO: Each service in its own repo.
//   PROS: Independent CI/CD, clear ownership, isolated
//   CONS: Cross-service changes = multiple PRs, version conflicts
//
// CI/CD STAGES:
//   1. Lint (golangci-lint)
//   2. Unit tests
//   3. Build Docker image
//   4. Integration tests (testcontainers)
//   5. Contract tests (Pact)
//   6. Push image to registry
//   7. Deploy to staging
//   8. Smoke tests
//   9. Canary deploy to production
//   10. Monitor

func main() {
	// =========================================================================
	// DEMO 1: Optimal Go Dockerfile
	// =========================================================================
	fmt.Println("=== DOCKER — Optimal Go Dockerfile ===")

	df := OptimalGoDockerfile()
	fmt.Println("  # Multi-stage Dockerfile")
	for _, stage := range df.Stages {
		fmt.Printf("  FROM %s AS %s\n", stage.BaseImage, stage.Name)
		for _, inst := range stage.Instructions {
			fmt.Printf("    %s\n", inst)
		}
	}
	fmt.Println("\n  Result: ~10-20MB image (vs ~800MB without multi-stage)")
	fmt.Println("  Security: non-root user, no shell, no package manager")

	// =========================================================================
	// DEMO 2: Kubernetes Manifests
	// =========================================================================
	fmt.Println("\n=== KUBERNETES MANIFESTS ===")

	manifests := OrderServiceK8sManifests()
	for _, m := range manifests {
		fmt.Printf("  %s: %s (ns: %s)\n", m.Kind, m.Name, m.Namespace)
		for k, v := range m.Spec {
			fmt.Printf("    %s: %v\n", k, v)
		}
		fmt.Println()
	}

	// =========================================================================
	// DEMO 3: Deployment Strategies Comparison
	// =========================================================================
	fmt.Println("=== DEPLOYMENT STRATEGIES ===")

	for _, s := range strategies {
		fmt.Printf("  %s:\n", s.Name)
		fmt.Printf("    %s\n", s.Description)
		fmt.Printf("    Rollback: %s | Cost: %s | Risk: %s\n",
			s.RollbackTime, s.ResourceCost, s.Risk)
		fmt.Printf("    Best for: %s\n\n", s.BestFor)
	}

	// =========================================================================
	// DEMO 4: Canary Deployment Simulation
	// =========================================================================
	fmt.Println("=== CANARY DEPLOYMENT ===")
	canary := NewCanaryDeployment("order-service", "v1.2.3", "v1.3.0")
	canary.Simulate()

	// =========================================================================
	// DEMO 5: Service Mesh Comparison
	// =========================================================================
	fmt.Println("\n=== SERVICE MESH COMPARISON ===")
	for _, mesh := range meshComparison {
		fmt.Printf("  %s (proxy: %s, CPU: %s, Mem: %s)\n",
			mesh.Mesh, mesh.ProxyType, mesh.ProxyCPU, mesh.ProxyMemory)
		fmt.Printf("    Features: %v\n", mesh.Features)
	}

	// =========================================================================
	// DEMO 6: GitOps Pipeline
	// =========================================================================
	fmt.Println("\n=== GITOPS FLOW ===")
	for _, step := range gitOpsSteps {
		fmt.Printf("  %d. [%s] %s\n     → %s\n", step.Step, step.Tool, step.Action, step.Description)
	}

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("\n=== DEPLOYMENT STACK ===")
	fmt.Println("┌──────────────────────────┬──────────────────────────────────────┐")
	fmt.Println("│ Layer                    │ Tools                                │")
	fmt.Println("├──────────────────────────┼──────────────────────────────────────┤")
	fmt.Println("│ Containerization         │ Docker, Podman, Buildpacks           │")
	fmt.Println("│ Orchestration            │ Kubernetes, Nomad, ECS               │")
	fmt.Println("│ Package Management       │ Helm, Kustomize                      │")
	fmt.Println("│ Service Mesh             │ Istio, Linkerd, Consul Connect       │")
	fmt.Println("│ CI/CD                    │ GitHub Actions, GitLab CI, ArgoCD    │")
	fmt.Println("│ GitOps                   │ ArgoCD, FluxCD                       │")
	fmt.Println("│ Infrastructure as Code   │ Terraform, Pulumi, Crossplane        │")
	fmt.Println("│ Secrets                  │ Vault, AWS Secrets Manager, SOPS     │")
	fmt.Println("│ Registry                 │ ECR, GCR, Docker Hub, Harbor         │")
	fmt.Println("└──────────────────────────┴──────────────────────────────────────┘")
}
