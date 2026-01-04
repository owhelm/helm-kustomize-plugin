package main

import "helm.sh/helm/v4/pkg/postrenderer"

// Compile-time interface compliance check for KustomizePostRenderer.
var _ postrenderer.PostRenderer = (*KustomizePostRenderer)(nil)
