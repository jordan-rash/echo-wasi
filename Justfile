just := env_var_or_default("JUST", just_executable())
wasm_tools := env_var_or_default("WASM_TOOLS", "wasm-tools")
wit_deps := env_var_or_default("WIT_DEPS", "wit-deps")
wash := env_var_or_default("WASH", "wash")
tinygo := env_var_or_default("TINYGO", "tinygo")
wit_bindgen := env_var_or_default("WIT_BINDGEN", "wit-bindgen")

actorid := env_var_or_default("ACTOR_ID", "")
hostid := env_var_or_default("HOST_ID", "")

expected_wasm_path := "./build/echo-wasi.wasm"
expected_wasm_embed_path := "./build/echo-wasi.embed.wasm"
expected_wasm_component_path := "./build/echo-wasi.component.wasm"
wasm_preview2_output_path := "./wasi_snapshot_preview1.command.wasm"

_default:
  {{just}} --list

@versions:
  {{tinygo}} version
  {{wash}} --version
  {{wasm_tools}} --version
  {{wit_bindgen}} --version


build:
  mkdir -p build
  go generate
  {{tinygo}} build -target=wasi -scheduler=none -o {{expected_wasm_path}} .
  {{wasm_tools}} component embed --world echo ./wit {{expected_wasm_path}} -o {{expected_wasm_embed_path}}
  {{wasm_tools}} component new {{expected_wasm_embed_path}} --adapt wasi_snapshot_preview1={{wasm_preview2_output_path}} -o {{expected_wasm_component_path}}
  {{wash}} claims sign --name echo-wasi {{expected_wasm_component_path}} -l -q -v 0.1.0 -r 0

clean:
  rm -rf ./build

signed_actor_path := absolute_path("./build/echo-wasi.component_s.wasm")
start_actor:
  {{wash}} start actor file://{{signed_actor_path}} --host-id {{hostid}}

restart_actor:
  {{wash}} stop actor {{hostid}} {{actorid}}
  {{wash}} drain lib
  {{wash}} start actor file://{{signed_actor_path}} --host-id {{hostid}}
