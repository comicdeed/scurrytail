{ lib
, buildGo127Module
, makeWrapper
, systemd
, util-linux
, xdg-utils
}:

buildGo127Module {
  pname = "scurrytail";
  version = "0.3.0";

  src = lib.cleanSource ./..;

  vendorHash = "sha256-ekjpCaGVEw8J0r3XqvTxde3Tbr7nR8ITEvxLD/4tZRE=";

  env.CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=0.3.0"
  ];

  nativeBuildInputs = [ makeWrapper ];

  postInstall = ''
    wrapProgram $out/bin/scurrytail \
      --prefix PATH : ${lib.makeBinPath [ systemd util-linux xdg-utils ]}
  '';

  meta = with lib; {
    description = "A nimble Linux system tray applet for tailing fast-moving kernel logs and catching errors in real time";
    homepage = "https://github.com/comicdeed/scurrytail";
    license = licenses.mit;
    mainProgram = "scurrytail";
    platforms = platforms.linux;
  };
}
