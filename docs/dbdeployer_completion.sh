# bash completion for dbdeployer                           -*- shell-script -*-

__dbdeployer_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__dbdeployer_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__dbdeployer_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__dbdeployer_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__dbdeployer_handle_go_custom_completion()
{
    __dbdeployer_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly dbdeployer allows to handle aliases
    args=("${words[@]:1}")
    requestComp="${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __dbdeployer_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __dbdeployer_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __dbdeployer_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __dbdeployer_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __dbdeployer_debug "${FUNCNAME[0]}: the completions are: ${out[*]}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __dbdeployer_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __dbdeployer_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __dbdeployer_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out[*]}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __dbdeployer_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subdir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out[0]}")
        if [ -n "$subdir" ]; then
            __dbdeployer_debug "Listing directories in $subdir"
            __dbdeployer_handle_subdirs_in_dir_flag "$subdir"
        else
            __dbdeployer_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out[*]}" -- "$cur")
    fi
}

__dbdeployer_handle_reply()
{
    __dbdeployer_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __dbdeployer_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION:-}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi

            if [[ -z "${flag_parsing_disabled}" ]]; then
                # If flag parsing is enabled, we have completed the flags and can return.
                # If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
                # to possibly call handle_go_custom_completion.
                return 0;
            fi
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __dbdeployer_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __dbdeployer_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        if declare -F __dbdeployer_custom_func >/dev/null; then
            # try command name qualified custom func
            __dbdeployer_custom_func
        else
            # otherwise fall back to unqualified for compatibility
            declare -F __custom_func >/dev/null && __custom_func
        fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__dbdeployer_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__dbdeployer_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__dbdeployer_handle_flag()
{
    __dbdeployer_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue=""
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __dbdeployer_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __dbdeployer_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __dbdeployer_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __dbdeployer_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        __dbdeployer_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__dbdeployer_handle_noun()
{
    __dbdeployer_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __dbdeployer_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __dbdeployer_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__dbdeployer_handle_command()
{
    __dbdeployer_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_dbdeployer_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __dbdeployer_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__dbdeployer_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __dbdeployer_handle_reply
        return
    fi
    __dbdeployer_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __dbdeployer_handle_flag
    elif __dbdeployer_contains_word "${words[c]}" "${commands[@]}"; then
        __dbdeployer_handle_command
    elif [[ $c -eq 0 ]]; then
        __dbdeployer_handle_command
    elif __dbdeployer_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __dbdeployer_handle_command
        else
            __dbdeployer_handle_noun
        fi
    else
        __dbdeployer_handle_noun
    fi
    __dbdeployer_handle_word
}

_dbdeployer_admin_capabilities()
{
    last_command="dbdeployer_admin_capabilities"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_admin_lock()
{
    last_command="dbdeployer_admin_lock"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_admin_remove-default()
{
    last_command="dbdeployer_admin_remove-default"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--default-sandbox-executable=")
    two_word_flags+=("--default-sandbox-executable")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_admin_set-default()
{
    last_command="dbdeployer_admin_set-default"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--default-sandbox-executable=")
    two_word_flags+=("--default-sandbox-executable")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_admin_unlock()
{
    last_command="dbdeployer_admin_unlock"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_admin_upgrade()
{
    last_command="dbdeployer_admin_upgrade"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_admin()
{
    last_command="dbdeployer_admin"

    command_aliases=()

    commands=()
    commands+=("capabilities")
    commands+=("lock")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("preserve")
        aliashash["preserve"]="lock"
    fi
    commands+=("remove-default")
    commands+=("set-default")
    commands+=("unlock")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("unpreserve")
        aliashash["unpreserve"]="unlock"
    fi
    commands+=("upgrade")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_cookbook_create()
{
    last_command="dbdeployer_cookbook_create"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_cookbook_list()
{
    last_command="dbdeployer_cookbook_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--sort-by=")
    two_word_flags+=("--sort-by")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_cookbook_show()
{
    last_command="dbdeployer_cookbook_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--raw")
    local_nonpersistent_flags+=("--raw")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_cookbook()
{
    last_command="dbdeployer_cookbook"

    command_aliases=()

    commands=()
    commands+=("create")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("make")
        aliashash["make"]="create"
    fi
    commands+=("list")
    commands+=("show")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_data-load_export()
{
    last_command="dbdeployer_data-load_export"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_data-load_get()
{
    last_command="dbdeployer_data-load_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--overwrite")
    local_nonpersistent_flags+=("--overwrite")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_data-load_import()
{
    last_command="dbdeployer_data-load_import"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_data-load_list()
{
    last_command="dbdeployer_data-load_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--full-info")
    local_nonpersistent_flags+=("--full-info")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_data-load_reset()
{
    last_command="dbdeployer_data-load_reset"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_data-load_show()
{
    last_command="dbdeployer_data-load_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--full-info")
    local_nonpersistent_flags+=("--full-info")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_data-load()
{
    last_command="dbdeployer_data-load"

    command_aliases=()

    commands=()
    commands+=("export")
    commands+=("get")
    commands+=("import")
    commands+=("list")
    commands+=("reset")
    commands+=("show")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_enable-bash-completion()
{
    last_command="dbdeployer_defaults_enable-bash-completion"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--completion-file=")
    two_word_flags+=("--completion-file")
    flags+=("--remote")
    flags+=("--remote-url=")
    two_word_flags+=("--remote-url")
    flags+=("--run-it")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_export()
{
    last_command="dbdeployer_defaults_export"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_flag-aliases()
{
    last_command="dbdeployer_defaults_flag-aliases"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_load()
{
    last_command="dbdeployer_defaults_load"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_reset()
{
    last_command="dbdeployer_defaults_reset"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_show()
{
    last_command="dbdeployer_defaults_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--camel-case")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_store()
{
    last_command="dbdeployer_defaults_store"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_templates_describe()
{
    last_command="dbdeployer_defaults_templates_describe"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--with-contents")
    local_nonpersistent_flags+=("--with-contents")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_templates_export()
{
    last_command="dbdeployer_defaults_templates_export"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_templates_import()
{
    last_command="dbdeployer_defaults_templates_import"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_templates_list()
{
    last_command="dbdeployer_defaults_templates_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--simple")
    flags+=("-s")
    local_nonpersistent_flags+=("--simple")
    local_nonpersistent_flags+=("-s")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_templates_reset()
{
    last_command="dbdeployer_defaults_templates_reset"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_templates_show()
{
    last_command="dbdeployer_defaults_templates_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_templates()
{
    last_command="dbdeployer_defaults_templates"

    command_aliases=()

    commands=()
    commands+=("describe")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("descr")
        aliashash["descr"]="describe"
        command_aliases+=("struct")
        aliashash["struct"]="describe"
        command_aliases+=("structure")
        aliashash["structure"]="describe"
    fi
    commands+=("export")
    commands+=("import")
    commands+=("list")
    commands+=("reset")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("remove")
        aliashash["remove"]="reset"
    fi
    commands+=("show")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults_update()
{
    last_command="dbdeployer_defaults_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults()
{
    last_command="dbdeployer_defaults"

    command_aliases=()

    commands=()
    commands+=("enable-bash-completion")
    commands+=("export")
    commands+=("flag-aliases")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("aliases")
        aliashash["aliases"]="flag-aliases"
        command_aliases+=("option-aliases")
        aliashash["option-aliases"]="flag-aliases"
    fi
    commands+=("load")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("import")
        aliashash["import"]="load"
    fi
    commands+=("reset")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("remove")
        aliashash["remove"]="reset"
    fi
    commands+=("show")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("list")
        aliashash["list"]="show"
    fi
    commands+=("store")
    commands+=("templates")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("templ")
        aliashash["templ"]="templates"
        command_aliases+=("template")
        aliashash["template"]="templates"
        command_aliases+=("tmpl")
        aliashash["tmpl"]="templates"
    fi
    commands+=("update")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_delete()
{
    last_command="dbdeployer_delete"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--concurrent")
    local_nonpersistent_flags+=("--concurrent")
    flags+=("--confirm")
    local_nonpersistent_flags+=("--confirm")
    flags+=("--skip-confirm")
    local_nonpersistent_flags+=("--skip-confirm")
    flags+=("--use-stop")
    local_nonpersistent_flags+=("--use-stop")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_delete-binaries()
{
    last_command="dbdeployer_delete-binaries"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--skip-confirm")
    local_nonpersistent_flags+=("--skip-confirm")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_deploy_multiple()
{
    last_command="dbdeployer_deploy_multiple"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--nodes=")
    two_word_flags+=("--nodes")
    two_word_flags+=("-n")
    flags+=("--base-port=")
    two_word_flags+=("--base-port")
    flags+=("--base-server-id=")
    two_word_flags+=("--base-server-id")
    flags+=("--binary-version=")
    two_word_flags+=("--binary-version")
    flags+=("--bind-address=")
    two_word_flags+=("--bind-address")
    flags+=("--client-from=")
    two_word_flags+=("--client-from")
    flags+=("--concurrent")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--custom-mysqld=")
    two_word_flags+=("--custom-mysqld")
    flags+=("--custom-role-extra=")
    two_word_flags+=("--custom-role-extra")
    flags+=("--custom-role-name=")
    two_word_flags+=("--custom-role-name")
    flags+=("--custom-role-privileges=")
    two_word_flags+=("--custom-role-privileges")
    flags+=("--custom-role-target=")
    two_word_flags+=("--custom-role-target")
    flags+=("--db-password=")
    two_word_flags+=("--db-password")
    two_word_flags+=("-p")
    flags+=("--db-user=")
    two_word_flags+=("--db-user")
    two_word_flags+=("-u")
    flags+=("--default-role=")
    two_word_flags+=("--default-role")
    flags+=("--defaults=")
    two_word_flags+=("--defaults")
    flags+=("--disable-mysqlx")
    flags+=("--enable-admin-address")
    flags+=("--enable-general-log")
    flags+=("--enable-mysqlx")
    flags+=("--expose-dd-tables")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--flavor-in-prompt")
    flags+=("--force")
    flags+=("--gtid")
    flags+=("--history-dir=")
    two_word_flags+=("--history-dir")
    flags+=("--init-general-log")
    flags+=("--init-options=")
    two_word_flags+=("--init-options")
    two_word_flags+=("-i")
    flags+=("--keep-server-uuid")
    flags+=("--log-directory=")
    two_word_flags+=("--log-directory")
    flags+=("--log-sb-operations")
    flags+=("--my-cnf-file=")
    two_word_flags+=("--my-cnf-file")
    flags+=("--my-cnf-options=")
    two_word_flags+=("--my-cnf-options")
    two_word_flags+=("-c")
    flags+=("--native-auth-plugin")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-as-server-id")
    flags+=("--post-grants-sql=")
    two_word_flags+=("--post-grants-sql")
    flags+=("--post-grants-sql-file=")
    two_word_flags+=("--post-grants-sql-file")
    flags+=("--pre-grants-sql=")
    two_word_flags+=("--pre-grants-sql")
    flags+=("--pre-grants-sql-file=")
    two_word_flags+=("--pre-grants-sql-file")
    flags+=("--remote-access=")
    two_word_flags+=("--remote-access")
    flags+=("--repl-crash-safe")
    flags+=("--rpl-password=")
    two_word_flags+=("--rpl-password")
    flags+=("--rpl-user=")
    two_word_flags+=("--rpl-user")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-directory=")
    two_word_flags+=("--sandbox-directory")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")
    flags+=("--skip-load-grants")
    flags+=("--skip-report-host")
    flags+=("--skip-report-port")
    flags+=("--skip-start")
    flags+=("--socket-in-datadir")
    flags+=("--task-user=")
    two_word_flags+=("--task-user")
    flags+=("--task-user-role=")
    two_word_flags+=("--task-user-role")
    flags+=("--use-template=")
    two_word_flags+=("--use-template")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_deploy_replication()
{
    last_command="dbdeployer_deploy_replication"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--change-master-options=")
    two_word_flags+=("--change-master-options")
    flags+=("--master-ip=")
    two_word_flags+=("--master-ip")
    flags+=("--master-list=")
    two_word_flags+=("--master-list")
    flags+=("--ndb-nodes=")
    two_word_flags+=("--ndb-nodes")
    flags+=("--nodes=")
    two_word_flags+=("--nodes")
    two_word_flags+=("-n")
    flags+=("--read-only-slaves")
    flags+=("--repl-history-dir")
    flags+=("--semi-sync")
    flags+=("--single-primary")
    flags+=("--slave-list=")
    two_word_flags+=("--slave-list")
    flags+=("--super-read-only-slaves")
    flags+=("--topology=")
    two_word_flags+=("--topology")
    two_word_flags+=("-t")
    flags+=("--base-port=")
    two_word_flags+=("--base-port")
    flags+=("--base-server-id=")
    two_word_flags+=("--base-server-id")
    flags+=("--binary-version=")
    two_word_flags+=("--binary-version")
    flags+=("--bind-address=")
    two_word_flags+=("--bind-address")
    flags+=("--client-from=")
    two_word_flags+=("--client-from")
    flags+=("--concurrent")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--custom-mysqld=")
    two_word_flags+=("--custom-mysqld")
    flags+=("--custom-role-extra=")
    two_word_flags+=("--custom-role-extra")
    flags+=("--custom-role-name=")
    two_word_flags+=("--custom-role-name")
    flags+=("--custom-role-privileges=")
    two_word_flags+=("--custom-role-privileges")
    flags+=("--custom-role-target=")
    two_word_flags+=("--custom-role-target")
    flags+=("--db-password=")
    two_word_flags+=("--db-password")
    two_word_flags+=("-p")
    flags+=("--db-user=")
    two_word_flags+=("--db-user")
    two_word_flags+=("-u")
    flags+=("--default-role=")
    two_word_flags+=("--default-role")
    flags+=("--defaults=")
    two_word_flags+=("--defaults")
    flags+=("--disable-mysqlx")
    flags+=("--enable-admin-address")
    flags+=("--enable-general-log")
    flags+=("--enable-mysqlx")
    flags+=("--expose-dd-tables")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--flavor-in-prompt")
    flags+=("--force")
    flags+=("--gtid")
    flags+=("--history-dir=")
    two_word_flags+=("--history-dir")
    flags+=("--init-general-log")
    flags+=("--init-options=")
    two_word_flags+=("--init-options")
    two_word_flags+=("-i")
    flags+=("--keep-server-uuid")
    flags+=("--log-directory=")
    two_word_flags+=("--log-directory")
    flags+=("--log-sb-operations")
    flags+=("--my-cnf-file=")
    two_word_flags+=("--my-cnf-file")
    flags+=("--my-cnf-options=")
    two_word_flags+=("--my-cnf-options")
    two_word_flags+=("-c")
    flags+=("--native-auth-plugin")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-as-server-id")
    flags+=("--post-grants-sql=")
    two_word_flags+=("--post-grants-sql")
    flags+=("--post-grants-sql-file=")
    two_word_flags+=("--post-grants-sql-file")
    flags+=("--pre-grants-sql=")
    two_word_flags+=("--pre-grants-sql")
    flags+=("--pre-grants-sql-file=")
    two_word_flags+=("--pre-grants-sql-file")
    flags+=("--remote-access=")
    two_word_flags+=("--remote-access")
    flags+=("--repl-crash-safe")
    flags+=("--rpl-password=")
    two_word_flags+=("--rpl-password")
    flags+=("--rpl-user=")
    two_word_flags+=("--rpl-user")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-directory=")
    two_word_flags+=("--sandbox-directory")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")
    flags+=("--skip-load-grants")
    flags+=("--skip-report-host")
    flags+=("--skip-report-port")
    flags+=("--skip-start")
    flags+=("--socket-in-datadir")
    flags+=("--task-user=")
    two_word_flags+=("--task-user")
    flags+=("--task-user-role=")
    two_word_flags+=("--task-user-role")
    flags+=("--use-template=")
    two_word_flags+=("--use-template")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_deploy_single()
{
    last_command="dbdeployer_deploy_single"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--master")
    flags+=("--prompt=")
    two_word_flags+=("--prompt")
    flags+=("--server-id=")
    two_word_flags+=("--server-id")
    flags+=("--base-port=")
    two_word_flags+=("--base-port")
    flags+=("--base-server-id=")
    two_word_flags+=("--base-server-id")
    flags+=("--binary-version=")
    two_word_flags+=("--binary-version")
    flags+=("--bind-address=")
    two_word_flags+=("--bind-address")
    flags+=("--client-from=")
    two_word_flags+=("--client-from")
    flags+=("--concurrent")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--custom-mysqld=")
    two_word_flags+=("--custom-mysqld")
    flags+=("--custom-role-extra=")
    two_word_flags+=("--custom-role-extra")
    flags+=("--custom-role-name=")
    two_word_flags+=("--custom-role-name")
    flags+=("--custom-role-privileges=")
    two_word_flags+=("--custom-role-privileges")
    flags+=("--custom-role-target=")
    two_word_flags+=("--custom-role-target")
    flags+=("--db-password=")
    two_word_flags+=("--db-password")
    two_word_flags+=("-p")
    flags+=("--db-user=")
    two_word_flags+=("--db-user")
    two_word_flags+=("-u")
    flags+=("--default-role=")
    two_word_flags+=("--default-role")
    flags+=("--defaults=")
    two_word_flags+=("--defaults")
    flags+=("--disable-mysqlx")
    flags+=("--enable-admin-address")
    flags+=("--enable-general-log")
    flags+=("--enable-mysqlx")
    flags+=("--expose-dd-tables")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--flavor-in-prompt")
    flags+=("--force")
    flags+=("--gtid")
    flags+=("--history-dir=")
    two_word_flags+=("--history-dir")
    flags+=("--init-general-log")
    flags+=("--init-options=")
    two_word_flags+=("--init-options")
    two_word_flags+=("-i")
    flags+=("--keep-server-uuid")
    flags+=("--log-directory=")
    two_word_flags+=("--log-directory")
    flags+=("--log-sb-operations")
    flags+=("--my-cnf-file=")
    two_word_flags+=("--my-cnf-file")
    flags+=("--my-cnf-options=")
    two_word_flags+=("--my-cnf-options")
    two_word_flags+=("-c")
    flags+=("--native-auth-plugin")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-as-server-id")
    flags+=("--post-grants-sql=")
    two_word_flags+=("--post-grants-sql")
    flags+=("--post-grants-sql-file=")
    two_word_flags+=("--post-grants-sql-file")
    flags+=("--pre-grants-sql=")
    two_word_flags+=("--pre-grants-sql")
    flags+=("--pre-grants-sql-file=")
    two_word_flags+=("--pre-grants-sql-file")
    flags+=("--remote-access=")
    two_word_flags+=("--remote-access")
    flags+=("--repl-crash-safe")
    flags+=("--rpl-password=")
    two_word_flags+=("--rpl-password")
    flags+=("--rpl-user=")
    two_word_flags+=("--rpl-user")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-directory=")
    two_word_flags+=("--sandbox-directory")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")
    flags+=("--skip-load-grants")
    flags+=("--skip-report-host")
    flags+=("--skip-report-port")
    flags+=("--skip-start")
    flags+=("--socket-in-datadir")
    flags+=("--task-user=")
    two_word_flags+=("--task-user")
    flags+=("--task-user-role=")
    two_word_flags+=("--task-user-role")
    flags+=("--use-template=")
    two_word_flags+=("--use-template")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_deploy()
{
    last_command="dbdeployer_deploy"

    command_aliases=()

    commands=()
    commands+=("multiple")
    commands+=("replication")
    commands+=("single")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--base-port=")
    two_word_flags+=("--base-port")
    flags+=("--base-server-id=")
    two_word_flags+=("--base-server-id")
    flags+=("--binary-version=")
    two_word_flags+=("--binary-version")
    flags+=("--bind-address=")
    two_word_flags+=("--bind-address")
    flags+=("--client-from=")
    two_word_flags+=("--client-from")
    flags+=("--concurrent")
    flags+=("--custom-mysqld=")
    two_word_flags+=("--custom-mysqld")
    flags+=("--custom-role-extra=")
    two_word_flags+=("--custom-role-extra")
    flags+=("--custom-role-name=")
    two_word_flags+=("--custom-role-name")
    flags+=("--custom-role-privileges=")
    two_word_flags+=("--custom-role-privileges")
    flags+=("--custom-role-target=")
    two_word_flags+=("--custom-role-target")
    flags+=("--db-password=")
    two_word_flags+=("--db-password")
    two_word_flags+=("-p")
    flags+=("--db-user=")
    two_word_flags+=("--db-user")
    two_word_flags+=("-u")
    flags+=("--default-role=")
    two_word_flags+=("--default-role")
    flags+=("--defaults=")
    two_word_flags+=("--defaults")
    flags+=("--disable-mysqlx")
    flags+=("--enable-admin-address")
    flags+=("--enable-general-log")
    flags+=("--enable-mysqlx")
    flags+=("--expose-dd-tables")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--flavor-in-prompt")
    flags+=("--force")
    flags+=("--gtid")
    flags+=("--history-dir=")
    two_word_flags+=("--history-dir")
    flags+=("--init-general-log")
    flags+=("--init-options=")
    two_word_flags+=("--init-options")
    two_word_flags+=("-i")
    flags+=("--keep-server-uuid")
    flags+=("--log-directory=")
    two_word_flags+=("--log-directory")
    flags+=("--log-sb-operations")
    flags+=("--my-cnf-file=")
    two_word_flags+=("--my-cnf-file")
    flags+=("--my-cnf-options=")
    two_word_flags+=("--my-cnf-options")
    two_word_flags+=("-c")
    flags+=("--native-auth-plugin")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-as-server-id")
    flags+=("--post-grants-sql=")
    two_word_flags+=("--post-grants-sql")
    flags+=("--post-grants-sql-file=")
    two_word_flags+=("--post-grants-sql-file")
    flags+=("--pre-grants-sql=")
    two_word_flags+=("--pre-grants-sql")
    flags+=("--pre-grants-sql-file=")
    two_word_flags+=("--pre-grants-sql-file")
    flags+=("--remote-access=")
    two_word_flags+=("--remote-access")
    flags+=("--repl-crash-safe")
    flags+=("--rpl-password=")
    two_word_flags+=("--rpl-password")
    flags+=("--rpl-user=")
    two_word_flags+=("--rpl-user")
    flags+=("--sandbox-directory=")
    two_word_flags+=("--sandbox-directory")
    flags+=("--skip-load-grants")
    flags+=("--skip-report-host")
    flags+=("--skip-report-port")
    flags+=("--skip-start")
    flags+=("--socket-in-datadir")
    flags+=("--task-user=")
    two_word_flags+=("--task-user")
    flags+=("--task-user-role=")
    two_word_flags+=("--task-user-role")
    flags+=("--use-template=")
    two_word_flags+=("--use-template")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_add()
{
    last_command="dbdeployer_downloads_add"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--OS=")
    two_word_flags+=("--OS")
    local_nonpersistent_flags+=("--OS")
    local_nonpersistent_flags+=("--OS=")
    flags+=("--arch=")
    two_word_flags+=("--arch")
    local_nonpersistent_flags+=("--arch")
    local_nonpersistent_flags+=("--arch=")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor=")
    flags+=("--minimal")
    local_nonpersistent_flags+=("--minimal")
    flags+=("--overwrite")
    local_nonpersistent_flags+=("--overwrite")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    local_nonpersistent_flags+=("--short-version")
    local_nonpersistent_flags+=("--short-version=")
    flags+=("--url=")
    two_word_flags+=("--url")
    local_nonpersistent_flags+=("--url")
    local_nonpersistent_flags+=("--url=")
    flags+=("--version=")
    two_word_flags+=("--version")
    local_nonpersistent_flags+=("--version")
    local_nonpersistent_flags+=("--version=")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_flag+=("--OS=")
    must_have_one_flag+=("--arch=")
    must_have_one_flag+=("--url=")
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_add-remote()
{
    last_command="dbdeployer_downloads_add-remote"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--minimal")
    local_nonpersistent_flags+=("--minimal")
    flags+=("--overwrite")
    local_nonpersistent_flags+=("--overwrite")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_export()
{
    last_command="dbdeployer_downloads_export"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--add-empty-item")
    local_nonpersistent_flags+=("--add-empty-item")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_get()
{
    last_command="dbdeployer_downloads_get"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--delete-after-unpack")
    local_nonpersistent_flags+=("--delete-after-unpack")
    flags+=("--dry-run")
    flags+=("--overwrite")
    flags+=("--prefix=")
    two_word_flags+=("--prefix")
    flags+=("--progress-step=")
    two_word_flags+=("--progress-step")
    local_nonpersistent_flags+=("--progress-step")
    local_nonpersistent_flags+=("--progress-step=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--retries-on-failure=")
    two_word_flags+=("--retries-on-failure")
    local_nonpersistent_flags+=("--retries-on-failure")
    local_nonpersistent_flags+=("--retries-on-failure=")
    flags+=("--shell")
    flags+=("--target-server=")
    two_word_flags+=("--target-server")
    flags+=("--unpack")
    local_nonpersistent_flags+=("--unpack")
    flags+=("--unpack-version=")
    two_word_flags+=("--unpack-version")
    flags+=("--verbosity=")
    two_word_flags+=("--verbosity")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_get-by-version()
{
    last_command="dbdeployer_downloads_get-by-version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--OS=")
    two_word_flags+=("--OS")
    local_nonpersistent_flags+=("--OS")
    local_nonpersistent_flags+=("--OS=")
    flags+=("--arch=")
    two_word_flags+=("--arch")
    local_nonpersistent_flags+=("--arch")
    local_nonpersistent_flags+=("--arch=")
    flags+=("--delete-after-unpack")
    local_nonpersistent_flags+=("--delete-after-unpack")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor=")
    flags+=("--guess-latest")
    local_nonpersistent_flags+=("--guess-latest")
    flags+=("--minimal")
    local_nonpersistent_flags+=("--minimal")
    flags+=("--newest")
    local_nonpersistent_flags+=("--newest")
    flags+=("--overwrite")
    flags+=("--prefix=")
    two_word_flags+=("--prefix")
    flags+=("--progress-step=")
    two_word_flags+=("--progress-step")
    local_nonpersistent_flags+=("--progress-step")
    local_nonpersistent_flags+=("--progress-step=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--retries-on-failure=")
    two_word_flags+=("--retries-on-failure")
    local_nonpersistent_flags+=("--retries-on-failure")
    local_nonpersistent_flags+=("--retries-on-failure=")
    flags+=("--shell")
    flags+=("--target-server=")
    two_word_flags+=("--target-server")
    flags+=("--unpack")
    local_nonpersistent_flags+=("--unpack")
    flags+=("--unpack-version=")
    two_word_flags+=("--unpack-version")
    flags+=("--verbosity=")
    two_word_flags+=("--verbosity")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_get-unpack()
{
    last_command="dbdeployer_downloads_get-unpack"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--delete-after-unpack")
    local_nonpersistent_flags+=("--delete-after-unpack")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--overwrite")
    flags+=("--prefix=")
    two_word_flags+=("--prefix")
    flags+=("--progress-step=")
    two_word_flags+=("--progress-step")
    local_nonpersistent_flags+=("--progress-step")
    local_nonpersistent_flags+=("--progress-step=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--retries-on-failure=")
    two_word_flags+=("--retries-on-failure")
    local_nonpersistent_flags+=("--retries-on-failure")
    local_nonpersistent_flags+=("--retries-on-failure=")
    flags+=("--shell")
    flags+=("--target-server=")
    two_word_flags+=("--target-server")
    flags+=("--unpack-version=")
    two_word_flags+=("--unpack-version")
    flags+=("--verbosity=")
    two_word_flags+=("--verbosity")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_import()
{
    last_command="dbdeployer_downloads_import"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--retries-on-failure=")
    two_word_flags+=("--retries-on-failure")
    local_nonpersistent_flags+=("--retries-on-failure")
    local_nonpersistent_flags+=("--retries-on-failure=")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_list()
{
    last_command="dbdeployer_downloads_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--OS=")
    two_word_flags+=("--OS")
    local_nonpersistent_flags+=("--OS")
    local_nonpersistent_flags+=("--OS=")
    flags+=("--arch=")
    two_word_flags+=("--arch")
    local_nonpersistent_flags+=("--arch")
    local_nonpersistent_flags+=("--arch=")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor=")
    flags+=("--show-url")
    local_nonpersistent_flags+=("--show-url")
    flags+=("--sort-by=")
    two_word_flags+=("--sort-by")
    local_nonpersistent_flags+=("--sort-by")
    local_nonpersistent_flags+=("--sort-by=")
    flags+=("--version=")
    two_word_flags+=("--version")
    local_nonpersistent_flags+=("--version")
    local_nonpersistent_flags+=("--version=")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_reset()
{
    last_command="dbdeployer_downloads_reset"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_show()
{
    last_command="dbdeployer_downloads_show"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads_tree()
{
    last_command="dbdeployer_downloads_tree"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--OS=")
    two_word_flags+=("--OS")
    local_nonpersistent_flags+=("--OS")
    local_nonpersistent_flags+=("--OS=")
    flags+=("--arch=")
    two_word_flags+=("--arch")
    local_nonpersistent_flags+=("--arch")
    local_nonpersistent_flags+=("--arch=")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor=")
    flags+=("--max-items=")
    two_word_flags+=("--max-items")
    local_nonpersistent_flags+=("--max-items")
    local_nonpersistent_flags+=("--max-items=")
    flags+=("--show-url")
    local_nonpersistent_flags+=("--show-url")
    flags+=("--version=")
    two_word_flags+=("--version")
    local_nonpersistent_flags+=("--version")
    local_nonpersistent_flags+=("--version=")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_flag+=("--flavor=")
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_downloads()
{
    last_command="dbdeployer_downloads"

    command_aliases=()

    commands=()
    commands+=("add")
    commands+=("add-remote")
    commands+=("export")
    commands+=("get")
    commands+=("get-by-version")
    commands+=("get-unpack")
    commands+=("import")
    commands+=("list")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("index")
        aliashash["index"]="list"
    fi
    commands+=("reset")
    commands+=("show")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("display")
        aliashash["display"]="show"
    fi
    commands+=("tree")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_export()
{
    last_command="dbdeployer_export"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--force-output-to-terminal")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_exec()
{
    last_command="dbdeployer_global_exec"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_metadata()
{
    last_command="dbdeployer_global_metadata"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_restart()
{
    last_command="dbdeployer_global_restart"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_start()
{
    last_command="dbdeployer_global_start"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_status()
{
    last_command="dbdeployer_global_status"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_stop()
{
    last_command="dbdeployer_global_stop"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_test()
{
    last_command="dbdeployer_global_test"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_test-replication()
{
    last_command="dbdeployer_global_test-replication"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global_use()
{
    last_command="dbdeployer_global_use"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--skip-library-check")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global()
{
    last_command="dbdeployer_global"

    command_aliases=()

    commands=()
    commands+=("exec")
    commands+=("metadata")
    commands+=("restart")
    commands+=("start")
    commands+=("status")
    commands+=("stop")
    commands+=("test")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("test-sb")
        aliashash["test-sb"]="test"
        command_aliases+=("test_sb")
        aliashash["test_sb"]="test"
    fi
    commands+=("test-replication")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("test_replication")
        aliashash["test_replication"]="test-replication"
    fi
    commands+=("use")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--name=")
    two_word_flags+=("--name")
    flags+=("--port=")
    two_word_flags+=("--port")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--short-version=")
    two_word_flags+=("--short-version")
    flags+=("--type=")
    two_word_flags+=("--type")
    flags+=("--verbose")
    flags+=("--version=")
    two_word_flags+=("--version")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_help()
{
    last_command="dbdeployer_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_dbdeployer_import_single()
{
    last_command="dbdeployer_import_single"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--client-from=")
    two_word_flags+=("--client-from")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-directory=")
    two_word_flags+=("--sandbox-directory")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_import()
{
    last_command="dbdeployer_import"

    command_aliases=()

    commands=()
    commands+=("single")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--client-from=")
    two_word_flags+=("--client-from")
    flags+=("--sandbox-directory=")
    two_word_flags+=("--sandbox-directory")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_info_defaults()
{
    last_command="dbdeployer_info_defaults"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--earliest")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_info_releases()
{
    last_command="dbdeployer_info_releases"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--limit=")
    two_word_flags+=("--limit")
    flags+=("--raw")
    flags+=("--stats")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--earliest")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_info_version()
{
    last_command="dbdeployer_info_version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--earliest")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_info()
{
    last_command="dbdeployer_info"

    command_aliases=()

    commands=()
    commands+=("defaults")
    commands+=("releases")
    commands+=("version")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--earliest")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_init()
{
    last_command="dbdeployer_init"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    flags+=("--skip-all-downloads")
    flags+=("--skip-shell-completion")
    flags+=("--skip-tarball-download")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_sandboxes()
{
    last_command="dbdeployer_sandboxes"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--by-date")
    local_nonpersistent_flags+=("--by-date")
    flags+=("--by-flavor")
    local_nonpersistent_flags+=("--by-flavor")
    flags+=("--by-version")
    local_nonpersistent_flags+=("--by-version")
    flags+=("--catalog")
    local_nonpersistent_flags+=("--catalog")
    flags+=("--flavor")
    local_nonpersistent_flags+=("--flavor")
    flags+=("--full-info")
    local_nonpersistent_flags+=("--full-info")
    flags+=("--header")
    local_nonpersistent_flags+=("--header")
    flags+=("--latest")
    local_nonpersistent_flags+=("--latest")
    flags+=("--oldest")
    local_nonpersistent_flags+=("--oldest")
    flags+=("--table")
    local_nonpersistent_flags+=("--table")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_unpack()
{
    last_command="dbdeployer_unpack"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--overwrite")
    flags+=("--prefix=")
    two_word_flags+=("--prefix")
    flags+=("--shell")
    flags+=("--target-server=")
    two_word_flags+=("--target-server")
    flags+=("--unpack-version=")
    two_word_flags+=("--unpack-version")
    flags+=("--verbosity=")
    two_word_flags+=("--verbosity")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_update()
{
    last_command="dbdeployer_update"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--OS=")
    two_word_flags+=("--OS")
    flags+=("--docs")
    local_nonpersistent_flags+=("--docs")
    flags+=("--dry-run")
    local_nonpersistent_flags+=("--dry-run")
    flags+=("--force-old-version")
    local_nonpersistent_flags+=("--force-old-version")
    flags+=("--new-path=")
    two_word_flags+=("--new-path")
    flags+=("--verbose")
    local_nonpersistent_flags+=("--verbose")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_usage()
{
    last_command="dbdeployer_usage"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_use()
{
    last_command="dbdeployer_use"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--ls")
    local_nonpersistent_flags+=("--ls")
    flags+=("--run=")
    two_word_flags+=("--run")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_versions()
{
    last_command="dbdeployer_versions"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--by-flavor")
    local_nonpersistent_flags+=("--by-flavor")
    flags+=("--flavor=")
    two_word_flags+=("--flavor")
    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_root_command()
{
    last_command="dbdeployer"

    command_aliases=()

    commands=()
    commands+=("admin")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("manage")
        aliashash["manage"]="admin"
    fi
    commands+=("cookbook")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("recipes")
        aliashash["recipes"]="cookbook"
        command_aliases+=("samples")
        aliashash["samples"]="cookbook"
    fi
    commands+=("data-load")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("load-data")
        aliashash["load-data"]="data-load"
    fi
    commands+=("defaults")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("config")
        aliashash["config"]="defaults"
    fi
    commands+=("delete")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("destroy")
        aliashash["destroy"]="delete"
        command_aliases+=("remove")
        aliashash["remove"]="delete"
    fi
    commands+=("delete-binaries")
    commands+=("deploy")
    commands+=("downloads")
    commands+=("export")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("dump")
        aliashash["dump"]="export"
    fi
    commands+=("global")
    commands+=("help")
    commands+=("import")
    commands+=("info")
    commands+=("init")
    commands+=("sandboxes")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("deployed")
        aliashash["deployed"]="sandboxes"
        command_aliases+=("installed")
        aliashash["installed"]="sandboxes"
    fi
    commands+=("unpack")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("expand")
        aliashash["expand"]="unpack"
        command_aliases+=("extract")
        aliashash["extract"]="unpack"
        command_aliases+=("inflate")
        aliashash["inflate"]="unpack"
        command_aliases+=("untar")
        aliashash["untar"]="unpack"
        command_aliases+=("unzip")
        aliashash["unzip"]="unpack"
    fi
    commands+=("update")
    commands+=("usage")
    commands+=("use")
    commands+=("versions")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("available")
        aliashash["available"]="versions"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    flags+=("--sandbox-binary=")
    two_word_flags+=("--sandbox-binary")
    flags+=("--sandbox-home=")
    two_word_flags+=("--sandbox-home")
    flags+=("--shell-path=")
    two_word_flags+=("--shell-path")
    flags+=("--skip-library-check")
    flags+=("--version")
    flags+=("-v")
    local_nonpersistent_flags+=("--version")
    local_nonpersistent_flags+=("-v")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_dbdeployer()
{
    local cur prev words cword split
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __dbdeployer_init_completion -n "=" || return
    fi

    local c=0
    local flag_parsing_disabled=
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("dbdeployer")
    local command_aliases=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function=""
    local last_command=""
    local nouns=()
    local noun_aliases=()

    __dbdeployer_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_dbdeployer dbdeployer
else
    complete -o default -o nospace -F __start_dbdeployer dbdeployer
fi

# ex: ts=4 sw=4 et filetype=sh
