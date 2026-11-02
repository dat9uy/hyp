# Sourced by launch.sh, doctor.sh and cleanup.sh.
# inputs_get KEY prints the header value of KEY in $HYP_SANDBOX_INPUTS.
inputs_get() {
	awk -v k="$1" 'NF==0 {exit} { i=index($0,"="); if (i>0 && substr($0,1,i-1)==k) {print substr($0,i+1); exit} }' "$HYP_SANDBOX_INPUTS"
}
die() { echo "skill-text-gate: $*" >&2; exit 1; }
