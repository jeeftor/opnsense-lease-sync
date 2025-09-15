{#
 # Copyright (c) 2024 DHCP AdGuard Sync
 # All rights reserved.
 #}

<script>
    $( document ).ready(function() {
        var data_get_map = {'frm_GeneralSettings':"/api/dhcpadguardsync/settings/get"};
        mapDataToFormUI(data_get_map).done(function(data){
            $('.selectpicker').selectpicker('refresh');
        });

        $("#saveAct").click(function(){
            saveFormToEndpoint(url="/api/dhcpadguardsync/settings/set", formid='frm_GeneralSettings',callback_ok=function(){
                $("#saveAct_progress").addClass("fa fa-spinner fa-pulse");
                ajaxCall(url="/api/dhcpadguardsync/service/reconfigure", sendData={}, callback=function(data,status) {
                    $("#saveAct_progress").removeClass("fa fa-spinner fa-pulse");
                    if (status != "success" || data['status'] != 'ok') {
                        BootstrapDialog.show({
                            type: BootstrapDialog.TYPE_WARNING,
                            title: "{{ lang._('Error reconfiguring DHCPAdGuardSync') }}",
                            message: data['status'],
                            draggable: true
                        });
                    }
                });
            });
        });

        $("#configtestAct").click(function(){
            $("#configtestAct_progress").addClass("fa fa-spinner fa-pulse");
            ajaxCall(url="/api/dhcpadguardsync/service/status", sendData={}, callback=function(data,status) {
                $("#configtestAct_progress").removeClass("fa fa-spinner fa-pulse");
                var message = data['status'];
                if (status != "success") {
                    message = "{{ lang._('Unable to run config test') }}";
                }
                BootstrapDialog.show({
                    type: data['status'] == 'running' ? BootstrapDialog.TYPE_INFO : BootstrapDialog.TYPE_WARNING,
                    title: "{{ lang._('Config test') }}",
                    message: message,
                    draggable: true
                });
            });
        });
    });
</script>

<div class="content-box" style="padding-bottom: 1.5em;">
    {{ partial("layout_partials/base_form",['fields':settingsForm,'id':'frm_GeneralSettings'])}}
    <div class="col-md-12">
        <hr />
        <button class="btn btn-primary" id="saveAct" type="button"><b>{{ lang._('Save') }}</b> <i id="saveAct_progress"></i></button>
        <button class="btn btn-info" id="configtestAct" type="button"><b>{{ lang._('Test Configuration') }}</b> <i id="configtestAct_progress"></i></button>
    </div>
</div>
