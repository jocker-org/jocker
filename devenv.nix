{ pkgs, lib, config, inputs, ... }:

{
  env.BUILDKIT_HOST = "docker-container://buildkit";
  packages = with pkgs; [ git gopls go buildkit jq mkdocs dive jsonnet jsonnet-bundler ];
  languages.go.enable = true;

  scripts.run-debug.exec = ''
    go run main.go debug-dump | buildctl debug dump-llb | jq
  '';

  scripts.run-build.exec = ''
    go run main.go debug-dump | buildctl build --local context=. --output type=image,name=docker.io/jocker/test,push=false
  '';

  scripts.run-publish.exec = ''
    docker build -t ghcr.io/jocker-org/jocker:latest -f Dockerfile --platform linux/amd64 . && docker push ghcr.io/jocker-org/jocker:latest
  '';
}
