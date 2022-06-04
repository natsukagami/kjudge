{ mkYarnPackage, ... }:
mkYarnPackage {
  name = "kjudge-frontend";
  # version = "0";
  src = ../frontend;
  packageJSON = ../frontend/package.json;
  yarnLock = ../frontend/yarn.lock;

  installPhase = ''
    DEST_DIR="$out/templates" yarn run --offline build:to
  '';
  distPhase = "true";
  configurePhase = "ln -s $node_modules node_modules";
}
