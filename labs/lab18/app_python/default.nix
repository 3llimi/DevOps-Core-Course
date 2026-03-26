{ pkgs ? import <nixpkgs> {} }:

let
  pythonEnv = pkgs.python3.withPackages (ps: with ps; [
    fastapi
    uvicorn
    pydantic
    starlette
    python-dotenv
    prometheus-client
  ]);

  cleanSrc = pkgs.lib.cleanSourceWith {
    src = ./.;
    filter = path: type:
      let
        base = builtins.baseNameOf path;
      in
        !(
          base == "venv" ||
          base == "__pycache__" ||
          base == ".pytest_cache" ||
          base == ".coverage" ||
          base == "app.log" ||
          base == "freeze1.txt" ||
          base == "freeze2.txt" ||
          base == "requirements-unpinned.txt" ||
          pkgs.lib.hasSuffix ".pyc" base
        );
  };
in
pkgs.stdenv.mkDerivation rec {
  pname = "devops-info-service";
  version = "1.0.0";
  src = cleanSrc;

  nativeBuildInputs = [ pkgs.makeWrapper ];

  installPhase = ''
    runHook preInstall
    mkdir -p $out/bin $out/app
    cp app.py $out/app/app.py
    makeWrapper ${pythonEnv}/bin/python $out/bin/devops-info-service \
      --add-flags "$out/app/app.py"
    runHook postInstall
  '';
}
