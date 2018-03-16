
export CATALOG=$HOME/.dbdeployer/sandboxes.json

dbdeployer_version=$(dbdeployer --version)
if [ -z "$dbdeployer_version" ]
then
    echo "dbdeployer not found"
    exit 1
fi

[ -z "$results_log" ] && export results_log=results-$(uname).txt

function start_timer {
    start=$(date)
    start_sec=$(date +%s)
    date > "$results_log"
}

function minutes_seconds {
    secs=$1
    elapsed_minutes=$((secs/60))
    remainder=$((secs-elapsed_minutes*60))
    echo "${elapsed_minutes}m:${remainder}s"
}

function stop_timer {
    stop=$(date)
    stop_sec=$(date +%s)
    elapsed=$(($stop_sec-$start_sec))
    echo "OS:  $(uname)"
    echo "OS:  $(uname)" >> "$results_log"
    echo "Started: $start"
    echo "Started: $start" >> "$results_log"
    echo "Ended  : $stop"
    echo "Ended  : $stop" >> "$results_log"
    echo "Elapsed: $elapsed seconds ($(minutes_seconds $elapsed))"
    echo "Elapsed: $elapsed seconds" >> "$results_log"
}

function show_catalog {
    if [ -f "$CATALOG" ]
    then
        cat "$CATALOG"
    fi
}

function count_catalog {
    show_catalog | grep destination | wc -l | tr -d ' '
}

function results {
    echo "#$*"
    echo "#$*" >> "$results_log"
    echo "dbdeployer sandboxes --catalog"
    echo "dbdeployer sandboxes --catalog" >> "$results_log"
    dbdeployer sandboxes --catalog
    dbdeployer sandboxes --catalog >> "$results_log"
    echo ""
    echo "" >> "$results_log"
    echo "catalog: $(count_catalog)" 
    echo "catalog: $(count_catalog)" >> "$results_log"
    if [ -n "$INTERACTIVE" ]
    then
        user_input
    fi
}

function ok_equal {
    label=$1
    value1=$2
    value2=$3
    if [ "$value1" == "$value2" ]
    then
        echo "ok - $label found '$value1' - expected: '$value2' "
        pass=$((pass+1))
    else
        echo "not ok - $label found '$value1' - expected: '$value2' "
        fail=$((fail+1))
    fi
    tests=$((tests+1))
}

function ok_contains {
    label=$1
    value1=$2
    value2=$3
    contains=$(echo "$value1" |grep "$value2")
    if [ -n "$contains" ]
    then
        echo "ok - $label - '$value1' contains '$value2' "
        pass=$((pass+1))
    else
        echo "not ok - $label - '$value1' does not contain '$value2' "
        fail=$((fail+1))
    fi
    tests=$((tests+1))
}

function ok {
    label=$1
    value=$2
    if [ -n "$value" ]
    then
        echo "ok - $label "
        pass=$((pass+1))
    else
        echo "not ok - $label "
        fail=$((fail+1))
    fi
    tests=$((tests+1))
}

function run {
    temp_stop_sec=$(date +%s)
    temp_elapsed=$(($temp_stop_sec-$start_sec))
    echo "+ $(date) (${temp_elapsed}s)"
    echo "+ $(date) (${temp_elapsed}s)" >> "$results_log"
    echo "# $*" >> "$results_log"
    (set -x
    $@
    )
    exit_code=$?
    echo $exit_code
    if [ "$exit_code" != "0" ]
    then
        exit $exit_code
    fi
}


