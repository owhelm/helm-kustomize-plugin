# experiment-helm-kustomize-plugin

This is an experiment to build a Kustomize plugin for Helm.

## Design

- The plugin supports the new Helm v4 plugin API via Extism
- It's a post-renderer plugin:
  - It expects the Helm chart to contain a special resource, which includes all the relevant files embedded inside of it
  - If it finds the special resource inside the chart
    - it extracts all the files contained in the special resource into a temporary folder
    - it removes the special resource from the chart output
    - it outputs the entire remaining contents of the chart into the `all.yaml` file
    - it updates the `kustomization.yaml` to reference the `all.yaml` under `resources` if it's not already referenced
    - it runs `kustomize` against the temporary folder and captures the output
    - it sends the output back to Helm
