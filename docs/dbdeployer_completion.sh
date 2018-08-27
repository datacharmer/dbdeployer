# bash completion for dbdeployer                           -*- shell-script -*-

__dbdeployer_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
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

__dbdeployer_handle_reply()
{
    __dbdeployer_debug "${FUNCNAME[0]}"
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
            COMPREPLY=( $(compgen -W "${allflags[*]}" -- "$cur") )
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
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi
            return 0;
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
        completions=("${must_have_one_noun[@]}")
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        COMPREPLY=( $(compgen -W "${noun_aliases[*]}" -- "$cur") )
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        declare -F __custom_func >/dev/null && __custom_func
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
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1
}

__dbdeployer_handle_flag()
{
    __dbdeployer_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
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
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if __dbdeployer_contains_word "${words[c]}" "${two_word_flags[@]}"; then
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
        if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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

    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_admin()
{
    last_command="dbdeployer_admin"

    command_aliases=()

    commands=()
    commands+=("lock")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("preserve")
        aliashash["preserve"]="lock"
    fi
    commands+=("unlock")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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

    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
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
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_defaults()
{
    last_command="dbdeployer_defaults"

    command_aliases=()

    commands=()
    commands+=("export")
    commands+=("load")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("import")
        aliashash["import"]="load"
    fi
    commands+=("reset")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("remove")
        aliashash["remove"]="reset"
    fi
    commands+=("show")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("list")
        aliashash["list"]="show"
    fi
    commands+=("store")
    commands+=("templates")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    two_word_flags+=("-n")
    flags+=("--base-port=")
    flags+=("--binary-version=")
    flags+=("--bind-address=")
    flags+=("--concurrent")
    flags+=("--config=")
    flags+=("--custom-mysqld=")
    flags+=("--db-password=")
    two_word_flags+=("-p")
    flags+=("--db-user=")
    two_word_flags+=("-u")
    flags+=("--defaults=")
    flags+=("--disable-mysqlx")
    flags+=("--enable-general-log")
    flags+=("--enable-mysqlx")
    flags+=("--expose-dd-tables")
    flags+=("--force")
    flags+=("--gtid")
    flags+=("--history-dir=")
    flags+=("--init-general-log")
    flags+=("--init-options=")
    two_word_flags+=("-i")
    flags+=("--keep-server-uuid")
    flags+=("--my-cnf-file=")
    flags+=("--my-cnf-options=")
    two_word_flags+=("-c")
    flags+=("--native-auth-plugin")
    flags+=("--port=")
    flags+=("--post-grants-sql=")
    flags+=("--post-grants-sql-file=")
    flags+=("--pre-grants-sql=")
    flags+=("--pre-grants-sql-file=")
    flags+=("--remote-access=")
    flags+=("--rpl-password=")
    flags+=("--rpl-user=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-directory=")
    flags+=("--sandbox-home=")
    flags+=("--skip-load-grants")
    flags+=("--skip-report-host")
    flags+=("--skip-report-port")
    flags+=("--skip-start")
    flags+=("--use-template=")

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

    flags+=("--master-ip=")
    flags+=("--master-list=")
    flags+=("--nodes=")
    two_word_flags+=("-n")
    flags+=("--repl-history-dir")
    flags+=("--semi-sync")
    flags+=("--single-primary")
    flags+=("--slave-list=")
    flags+=("--topology=")
    two_word_flags+=("-t")
    flags+=("--base-port=")
    flags+=("--binary-version=")
    flags+=("--bind-address=")
    flags+=("--concurrent")
    flags+=("--config=")
    flags+=("--custom-mysqld=")
    flags+=("--db-password=")
    two_word_flags+=("-p")
    flags+=("--db-user=")
    two_word_flags+=("-u")
    flags+=("--defaults=")
    flags+=("--disable-mysqlx")
    flags+=("--enable-general-log")
    flags+=("--enable-mysqlx")
    flags+=("--expose-dd-tables")
    flags+=("--force")
    flags+=("--gtid")
    flags+=("--history-dir=")
    flags+=("--init-general-log")
    flags+=("--init-options=")
    two_word_flags+=("-i")
    flags+=("--keep-server-uuid")
    flags+=("--my-cnf-file=")
    flags+=("--my-cnf-options=")
    two_word_flags+=("-c")
    flags+=("--native-auth-plugin")
    flags+=("--port=")
    flags+=("--post-grants-sql=")
    flags+=("--post-grants-sql-file=")
    flags+=("--pre-grants-sql=")
    flags+=("--pre-grants-sql-file=")
    flags+=("--remote-access=")
    flags+=("--rpl-password=")
    flags+=("--rpl-user=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-directory=")
    flags+=("--sandbox-home=")
    flags+=("--skip-load-grants")
    flags+=("--skip-report-host")
    flags+=("--skip-report-port")
    flags+=("--skip-start")
    flags+=("--use-template=")

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
    flags+=("--base-port=")
    flags+=("--binary-version=")
    flags+=("--bind-address=")
    flags+=("--concurrent")
    flags+=("--config=")
    flags+=("--custom-mysqld=")
    flags+=("--db-password=")
    two_word_flags+=("-p")
    flags+=("--db-user=")
    two_word_flags+=("-u")
    flags+=("--defaults=")
    flags+=("--disable-mysqlx")
    flags+=("--enable-general-log")
    flags+=("--enable-mysqlx")
    flags+=("--expose-dd-tables")
    flags+=("--force")
    flags+=("--gtid")
    flags+=("--history-dir=")
    flags+=("--init-general-log")
    flags+=("--init-options=")
    two_word_flags+=("-i")
    flags+=("--keep-server-uuid")
    flags+=("--my-cnf-file=")
    flags+=("--my-cnf-options=")
    two_word_flags+=("-c")
    flags+=("--native-auth-plugin")
    flags+=("--port=")
    flags+=("--post-grants-sql=")
    flags+=("--post-grants-sql-file=")
    flags+=("--pre-grants-sql=")
    flags+=("--pre-grants-sql-file=")
    flags+=("--remote-access=")
    flags+=("--rpl-password=")
    flags+=("--rpl-user=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-directory=")
    flags+=("--sandbox-home=")
    flags+=("--skip-load-grants")
    flags+=("--skip-report-host")
    flags+=("--skip-report-port")
    flags+=("--skip-start")
    flags+=("--use-template=")

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
    flags+=("--binary-version=")
    flags+=("--bind-address=")
    flags+=("--concurrent")
    flags+=("--custom-mysqld=")
    flags+=("--db-password=")
    two_word_flags+=("-p")
    flags+=("--db-user=")
    two_word_flags+=("-u")
    flags+=("--defaults=")
    flags+=("--disable-mysqlx")
    flags+=("--enable-general-log")
    flags+=("--enable-mysqlx")
    flags+=("--expose-dd-tables")
    flags+=("--force")
    flags+=("--gtid")
    flags+=("--history-dir=")
    flags+=("--init-general-log")
    flags+=("--init-options=")
    two_word_flags+=("-i")
    flags+=("--keep-server-uuid")
    flags+=("--my-cnf-file=")
    flags+=("--my-cnf-options=")
    two_word_flags+=("-c")
    flags+=("--native-auth-plugin")
    flags+=("--port=")
    flags+=("--post-grants-sql=")
    flags+=("--post-grants-sql-file=")
    flags+=("--pre-grants-sql=")
    flags+=("--pre-grants-sql-file=")
    flags+=("--remote-access=")
    flags+=("--rpl-password=")
    flags+=("--rpl-user=")
    flags+=("--sandbox-directory=")
    flags+=("--skip-load-grants")
    flags+=("--skip-report-host")
    flags+=("--skip-report-port")
    flags+=("--skip-start")
    flags+=("--use-template=")
    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_dbdeployer_global()
{
    last_command="dbdeployer_global"

    command_aliases=()

    commands=()
    commands+=("restart")
    commands+=("start")
    commands+=("status")
    commands+=("stop")
    commands+=("test")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("test-sb")
        aliashash["test-sb"]="test"
        command_aliases+=("test_sb")
        aliashash["test_sb"]="test"
    fi
    commands+=("test-replication")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("test_replication")
        aliashash["test_replication"]="test-replication"
    fi
    commands+=("use")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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

    flags+=("--catalog")
    local_nonpersistent_flags+=("--catalog")
    flags+=("--header")
    local_nonpersistent_flags+=("--header")
    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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

    flags+=("--prefix=")
    flags+=("--shell")
    flags+=("--target-server=")
    flags+=("--unpack-version=")
    flags+=("--verbosity=")
    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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

    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")

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
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("manage")
        aliashash["manage"]="admin"
    fi
    commands+=("defaults")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("config")
        aliashash["config"]="defaults"
    fi
    commands+=("delete")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("destroy")
        aliashash["destroy"]="delete"
        command_aliases+=("remove")
        aliashash["remove"]="delete"
    fi
    commands+=("deploy")
    commands+=("global")
    commands+=("sandboxes")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("deployed")
        aliashash["deployed"]="sandboxes"
        command_aliases+=("installed")
        aliashash["installed"]="sandboxes"
    fi
    commands+=("unpack")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
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
    commands+=("usage")
    commands+=("versions")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("available")
        aliashash["available"]="versions"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    flags+=("--sandbox-binary=")
    flags+=("--sandbox-home=")
    flags+=("--version")
    local_nonpersistent_flags+=("--version")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_dbdeployer()
{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __dbdeployer_init_completion -n "=" || return
    fi

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("dbdeployer")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local last_command
    local nouns=()

    __dbdeployer_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_dbdeployer dbdeployer
else
    complete -o default -o nospace -F __start_dbdeployer dbdeployer
fi

# ex: ts=4 sw=4 et filetype=sh
