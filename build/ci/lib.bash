#!/bin/bash

group() { echo "::group::$*"; }
groupend() { echo "::groupend::"; }
fail() { echo "::error::$*"; exit 1;}

cleanup_testing_environment()
{
    group "cleanup test system"
    suseconnect --clean || true  # ignore failures
    [ -f /etc/SUSEConnect ] && rm /etc/SUSEConnect
    groupend
}

check_required_env_vars_defined()
{
    local env_vars_to_check env_vars_missing

    env_vars_to_check=( ${@} )
    env_vars_missing=()

    group "verify required environment variables defined"
    echo "Expecting these environment variables to be set:"
    echo "${env_vars_to_check[@]}"
    for env_var in "${env_vars_to_check[@]}"; do
        if [ -z "${!env_var}" ]; then
            echo "ENV variable not found: ${env_var} is not set."
            env_vars_missing+=( "${env_var}" )
        fi
    done
    if (( ${#env_vars_missing[@]} )); then
        fail "Required environment variables not set: ${env_vars_missing[@]}"
    fi
    echo "All required environment variables provided"
    groupend
}

install_locally_built_suseconnect()
{
    local out_dir="$(readlink -e ${1})" src_dir bin_tools zypp_cmds rc_links systemd_libdir ruby_libdir

    src_dir="$(dirname "${out_dir}")"
    bin_tools=( suseconnect suse-uptime-tracker suseconnect-mcp )
    zypp_cmds=( zypper-migration zypper-search-packages )
    svc_cmds=( suseconnect-keepalive suse-uptime-tracker )
    systemd_libdir=/usr/lib/systemd/system
    ruby_libdir=/usr/lib64/ruby/vendor_ruby/3.4.0

    # install binary tools
    group "Install binary tools"
    for bt in "${bin_tools[@]}"
    do
        cp -pv "${out_dir}/${bt}" "/usr/bin/${bt}"
    done
    groupend

    # create binary tools symlinks
    group "Install binary tools symlinks"
    ln -fsv suseconnect /usr/bin/SUSEConnect
    ln -fsv ../bin/suseconnect /usr/sbin/SUSEConnect
    groupend

    # install zypper commands
    group "Install zypper commands"
    for zc in "${zypp_cmds[@]}"
    do
        cp -pv "${out_dir}/${zc}" "/usr/lib/zypper/${zc}"
    done
    groupend

    # install libsuseconnect
    group "Install shared libraries"
    for shlib in libsuseconnect.so
    do
        cp -pv "${out_dir}/${shlib}" "/usr/lib64/${shlib}"
    done
    groupend

    # install ruby bindings
    group "Install ruby bindings"
    mkdir -m 0755 -pv "${ruby_libdir}"
    cp -apv "${src_dir}/third_party/yast/lib/suse" "${ruby_libdir}/"
    groupend

    # optionally install service integration
    if [ -x /usr/sbin/service ]; then
        group "Install service links"
        for sc in "${svc_cmds[@]}"
        do
            ln -fsv service "/usr/sbin/rc${sc}"
        done
        groupend
    fi

    # optionally install systemd integration
    if [ -d "${systemd_libdir}" ]; then
        group "Install systemd integration"
        for sc in "${svc_cmds[@]}"
        do
            for svc_type in service timer
            do
                cp -pv "${src_dir}/build/packaging/${sc}.${svc_type}" "${systemd_libdir}"
            done
        done
        groupend
    fi
}

cleanup_orphaned_products()
{
    local products_to_check_for products_to_remove
    if (( $# > 0 )); then
        products_to_check_for=( "${@}" )
    else
        products_to_check_for=(
            sle-module-python3
            sle-module-server-applications
            sle-module-basesystem
        )
    fi
    products_to_remove=()

    for p in "${products_to_check_for[@]}"
    do
        if zypper --no-refresh info -t product "${p}" >/dev/null; then
            products_to_remove+=( "${p}" )
        fi
    done

    if (( ${#products_to_remove[@]} )); then
        group "Removing orphaned products: ${products_to_remove[@]}"
        zypper --no-refresh --non-interactive remove -t product "${products_to_remove[@]}"
        groupend
    fi
}