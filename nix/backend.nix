{
  templates,
  lib,
  writeText,
  buildGoModule,
  go,

  bash,
  gnugrep,
  sqlite,
  curl,
  ...
}:
let
  toYaml = lib.generators.toYAML { };

  # vendorHash = "sha256-TMxl3vbgOSB2dnPEP5FF7VVeyxSJRNtkDxUJBrrUZ78=";
  vendorHash = lib.fakeHash;

  # fileb0x configuration
  fileb0xConf = writeText "fileb0x.yaml" (toYaml {
    pkg = "static";
    dest = "./static";
    fmt = false;
    debug = false;

    compression.compress = true;

    custom = [
      {
        files = [
          "./assets/"
          "./templates/"
        ];
      }
    ];
  });
in
(buildGoModule.override { inherit go; }) {
  inherit vendorHash;

  src = ../.;
  tags = [ "production" ];

  name = "kjudge";

  nativeBuildInputs = [
    bash
    gnugrep
    sqlite
    curl
  ];

  # Generate sources
  preBuild = ''
    # Link templates
    cp -r ${templates}/templates .
    # Install tools
    bash scripts/install_tools.sh
    # Add GOPATH to PATH
    export PATH=$PATH:$(go env GOPATH)/bin
    # Generate fileb0x
    fileb0x ${fileb0xConf}
    # Generate sql models
    go run models/generate/main.go
  '';
}
