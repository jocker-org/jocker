{
  new:: function() {
    stages: [],
  },
  stage:: {
    new:: function(name, from) {
      name: name,
      from: from,
      steps: [],
    },
    withSteps:: function(steps) {
      steps: steps
    },
    withStepMixin:: function(step) {
      steps+: if std.isArray(step) then step else [step],
    },
    step:: {
      workdir:: function(dir) {type: "WORKDIR", "path": dir},
      run:: function(cmd) {type: "RUN", "command": cmd},
      user:: function(user) {type: "USER", "user": user},
      copy:: function(src, dst) {type:"COPY", src: src, dst: dst},
      copyFrom:: function(src, dst, from=null) {type:"COPY", from: from, src: src, dst: dst},
    },
  },
  withStage:: function(stage) {
    stages+: if std.isArray(stage) then stage else [stage],
  },
  withExcludes:: function(excludes) {
    excludes: if std.isArray(excludes) then excludes else [excludes],
  },
  withExcludeMixin:: function(exclude) {
    excludes+: if std.isArray(exclude) then exclude else [exclude],
  },
  withCmd:: function(cmd) {
    cmd: cmd,
  },
  withImage:: function(imageMetadata) {
    local baseImage = {
      "Env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
    },
    image: baseImage + imageMetadata
  },
}
