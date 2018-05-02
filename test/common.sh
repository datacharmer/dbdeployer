
export CATALOG=$HOME/.dbdeployer/sandboxes.json
export dash_line="# ----------------------------------------------------------------"
export dotted_line="# ................................................................"
export double_dash_line="# ================================================================"

if [ -n "$SKIP_DBDEPLOYER_CATALOG" ]
then
    echo "This test requires dbdeployer catalog to be enabled"
    echo "Unset the variable SKIP_DBDEPLOYER_CATALOG to continue"
    exit 1
fi

dbdeployer_version=$(dbdeployer --version)
if [ -z "$dbdeployer_version" ]
then
    echo "dbdeployer not found"
    exit 1
fi

[ -z "$results_log" ] && export results_log=results-$(uname).txt

function test_header {
    func_name=$1
    arg="$2"
    double_line=$3
    if [ -n "$double_line" ]
    then
        echo $double_dash_line
    else
        echo $dash_line
    fi
    echo "# $func_name $arg"
    if [ -n "$double_line" ]
    then
        echo $double_dash_line
    else
        echo $dash_line
    fi
}

function check_for_log_errors {
    label=$1
    skip_error_evaluation=$2
    for log_file in $(find $SANDBOX_HOME -name msandbox.err)
    do
        has_errors=$(grep -w ERROR $log_file| wc -l | tr -d ' \t' )
        if [ "$has_errors" != "0" ]
        then
            echo $dash_line
            echo "# called from: $label"
            echo "# log file:    $log_file"
            echo $dash_line
            grep -w ERROR $log_file
            echo $dash_line
        fi
        if [ -z "$skip_error_evaluation" ]
        then
            ok_equal "Health check for errors in log file $log_file" "$has_errors" "0"
        fi
    done
}

function check_for_exit {
    label=$1
    skip_log_check=$2
    echo "## >> Label for exit on demand : $label"
    check_for_log_errors $label $skip_log_check
    if [ "$exit_on_demand" == "$label" ]
    then
        echo "Exit on demand - label: $label"
        echo "pass: $pass"
        echo "fail: $fail"
        exit 0
    fi
}


function sandbox_num_ports {
    running_version=$1
    dir_name=$2
    version_path=$(echo $running_version| tr '.' '_')
    descr=$SANDBOX_HOME/$dir_name$version_path/sbdescription.json
    if [ ! -f $descr ]
    then
        echo 0
        return
    fi
    cat $descr | sed -n '/port/,/]/p' | grep '^\s*[0-9]\+' | wc -l | tr -d ' \t'
}

function start_timer {
    start=$(date)
    start_sec=$(date +%s)
    date > "$results_log"
}

function minutes_seconds {
    secs=$1
    if [ -z "$secs" ]
    then
        secs=0
    fi
    if [[ $secs -lt 60 ]]
    then
        echo "${secs}s"
        return
    fi
    elapsed_minutes=$((secs/60))
    remainder_sec=$((secs-elapsed_minutes*60))
    if [[ $elapsed_minutes -lt 60 ]]
    then
        printf "%dm:%02ds" ${elapsed_minutes} ${remainder_sec}
        return
    fi
    elapsed_hours=$((elapsed_minutes/60))
    remainder_min=$((elapsed_minutes-elapsed_hours*60))
    printf "%dh:%dm:%02ds" ${elapsed_hours} ${remainder_min} ${remainder_sec}
}

function stop_timer {
    stop_log=$1
    [ -z "$stop_log" ] && stop_log=$results_log
    stop=$(date)
    [ -z "$stop_sec" ] && stop_sec=$(date +%s)
    elapsed=$(($stop_sec-$start_sec))
    echo "OS:  $(uname)"
    echo "OS:  $(uname)" >> "$stop_log"
    echo "Started: $start"
    echo "Started: $start" >> "$stop_log"
    echo "Ended  : $stop"
    echo "Ended  : $stop" >> "$stop_log"
    echo "Elapsed: $elapsed seconds ($(minutes_seconds $elapsed))"
    echo "Elapsed: $elapsed seconds ($(minutes_seconds $elapsed))" >> "$stop_log"
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

function list_active_tests {
    echo "Enabled tests:"
    if [ -z "$skip_main_deployment_methods" ]
    then
        echo main_deployment_methods
    fi
    if [ -z "$skip_pre_post_operations" ]
    then
        echo pre_post_operations
    fi
    if [ -z "$skip_group_operations" ]
    then
        echo group_operations
    fi
    if [ -z "$skip_multi_source_operations" ]
    then
        echo multi_source_operations
    fi
    echo "Current test: $current_test"
    echo ""
    concurrency=no
    if [ -n "$RUN_CONCURRENTLY" ]
    then
        concurrency=yes
    fi
    echo "Runs concurrently: $concurrency"
    echo ""
}



function user_input {
    answer=""
    while [ "$answer" != "continue" ]
    do
        echo "Press ENTER to continue or choose among { s c q i o r u h t }"
        read answer
        case $answer in
            [cC])
                unset INTERACTIVE
                echo "Now running unattended"
                return
                ;;
            [qQ])
                echo "Interrupted at user's request"
                exit 0
                ;;
            [iI])
                echo inspecting
                show_catalog
                ;;
            [oO])
                echo counting
                count_catalog
                ;;
            [sS])
                echo show sandboxes
                dbdeployer sandboxes --catalog
                ;;
            [rR])
                echo "Enter global command to run"
                echo "Choose among : start restart stop status test test-replication"
                read cmd
                dbdeployer global $cmd
                if [ "$?" != "0" ]
                then
                    exit 1
                fi
                ;;
            [uU])
                echo "Enter query to run"
                read cmd
                dbdeployer global use "$cmd"
                if [ "$?" != "0" ]
                then
                    exit 1
                fi
                ;;
            [tT])
                list_active_tests
                ;;
            [hH])
                echo "Commands:"
                echo "c : continue (end interactivity)"
                echo "i : inspect sandbox catalog"
                echo "o : count sandbox instances"
                echo "q : quit the test immediately"
                echo "r : run 'dbdeployer global' command"
                echo "u : run 'dbdeployer global use' query"
                echo "s : show sandboxes"
                echo "t : list active tests"
                echo "h : display this help"
                ;;
            *)
                answer="continue"
        esac
    done
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

function ok_comparison {
    op=$1
    label=$2
    value1=$3
    value2=$4
    value3=$5
    unset success
    unset failure
    if [ -z "$value1"  -o -z "$value2" ]
    then
        echo "ok_$op: empty value passed"
        exit 1
    fi
    case $op in
        equal)
            expected="'$value2'"
            if [ -n "$value3" ]
            then
                expected="'$value2 or $value3'"
            fi
            if [ "$value1" == "$value2" -o "$value1" == "$value3" ]
            then
                success="ok - $label found '$value1' - expected: $expected "
            else
                failure="not ok - $label found '$value1' - expected: $expected "
            fi
            ;;
        not_equal)
            if [ "$value1" != "$value2" ]
            then
                success="ok - $label found '$value1' - expected: != $value2 "
            else
                failure="not ok - $label found '$value1' - expected: != $value2 "
            fi
            ;;
         greater)
            if [[ $value1 -gt $value2 ]]
            then
                success="ok - $label  '$value1' > '$value2' "
            else
                failure="not ok - $label  '$value1' not > '$value2' "
            fi
            ;;
        greater_equal)
            if [[ $value1 -ge $value2 ]]
            then
                success="ok - $label  '$value1' >= '$value2' "
            else
                failure="not ok - $label  '$value1' not >= '$value2' "
            fi
            ;;
        *)
            echo "Unsupported operation '$op'"
            exit 1
    esac
    if [ -n "$success" ]
    then
        echo $success
        pass=$((pass+1))
    elif [ -n "$failure" ]
    then
        echo $failure
        fail=$((fail+1))
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            echo "pass: $pass - fail: $fail"
            exit
        fi
    else
        echo "Neither success or failure detected"
        echo "op:     $op"
        echo "label:  $label"
        echo "value1: $value1 "
        echo "value2: $value2 "
        exit 1
    fi
    tests=$((tests+1))
}

function ok_equal {
    label=$1
    value1=$2
    value2=$3
    value3=$4
    ok_comparison equal "$label" "$value1" "$value2" "$value3"
}

function ok_not_equal {
    label=$1
    value1=$2
    value2=$3
    ok_comparison not_equal "$label" "$value1" "$value2"
}

function ok_greater {
    label="$1"
    value1=$2
    value2=$3
    ok_comparison greater "$label" "$value1" "$value2"
}

function ok_greater_equal {
    label="$1"
    value1=$2
    value2=$3
    ok_comparison greater_equal "$label" "$value1" "$value2"
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
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            echo "pass: $pass - fail: $fail"
            exit
        fi
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
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            echo "pass: $pass - fail: $fail"
            exit
        fi
    fi
    tests=$((tests+1))
}

function ok_generic_exists {
    wanted=$1
    label=$2
    op=$3
    if [ $op "$wanted" ]
    then
        echo "ok - $label $wanted exists"
        pass=$((pass+1))
    else
        echo "NOT OK - $label $wanted does not  exist"
        fail=$((fail+1))
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            echo "pass: $pass - fail: $fail"
            exit
        fi
    fi
    tests=$((tests+1))
}

function ok_generic_does_not_exist {
    wanted=$1
    label=$2
    op=$3
    if [ ! $op "$wanted" ]
    then
        echo "ok - $label $wanted does not exist"
        pass=$((pass+1))
    else
        echo "NOT OK - $label $wanted exists"
        fail=$((fail+1))
        if [ -n "$EXIT_ON_FAILURE" ]
        then
            echo "pass: $pass - fail: $fail"
            exit
        fi
    fi
    tests=$((tests+1))
}


function ok_dir_exists {
    dir=$1
    ok_generic_exists $dir directory -d
}

function ok_file_exists {
    filename=$1
    ok_generic_exists $filename "file" -f
}

function ok_executable_exists {
    filename=$1
    ok_generic_exists $filename "file" -x
}

function ok_executable_does_not_exist {
    filename=$1
    ok_generic_does_not_exist $filename "file" "-x"
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


