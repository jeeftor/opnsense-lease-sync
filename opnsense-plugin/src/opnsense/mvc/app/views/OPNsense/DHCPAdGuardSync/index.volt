{#
# Copyright (c) 2024 DHCP AdGuard Sync
# All rights reserved.
#}

<script>
    $( document ).ready(function() {
        var data_get_map = {'frm_GeneralSettings':"/api/dhcpadguardsync/settings/get"};
        mapDataToFormUI(data_get_map).done(function(data){
            formatTokenizersUI();
            $('.selectpicker').selectpicker('refresh');
        });

        // Save configuration
        $("#saveAct").click(function(){
            $("#saveAct_progress").addClass("fa fa-spinner fa-pulse");
            saveFormToEndpoint(url="/api/dhcpadguardsync/settings/set", formid='frm_GeneralSettings',callback_ok=function(){
                $("#saveAct_progress").removeClass("fa fa-spinner fa-pulse");
                $("#responseMsg").removeClass("hidden").html("{{ lang._('Settings saved successfully.') }}");
                // Auto-reconfigure service after save
                ajaxCall(url="/api/dhcpadguardsync/service/reconfigure", sendData={}, callback=function(data,status) {
                    if (status === "success" && data.status === 'ok') {
                        $("#responseMsg").html("{{ lang._('Settings saved and service reconfigured.') }}");
                    } else {
                        $("#responseMsg").html("{{ lang._('Settings saved but service reconfiguration failed.') }}");
                    }
                });
            }, callback_fail=function(){
                $("#saveAct_progress").removeClass("fa fa-spinner fa-pulse");
                $("#responseMsg").removeClass("hidden").addClass("alert-danger").html("{{ lang._('Error saving settings.') }}");
            });
        });

        // Service control functions
        function performServiceAction(action, button_id) {
            $("#" + button_id + "_progress").addClass("fa fa-spinner fa-pulse");
            ajaxCall(url="/api/dhcpadguardsync/service/" + action, sendData={}, callback=function(data,status) {
                $("#" + button_id + "_progress").removeClass("fa fa-spinner fa-pulse");
                if (status === "success") {
                    $("#serviceOutput").text(data.response || data.status);
                } else {
                    $("#serviceOutput").text("{{ lang._('Error performing') }} " + action);
                }
            });
        }

        $("#startAct").click(function(){ performServiceAction('start', 'startAct'); });
        $("#stopAct").click(function(){ performServiceAction('stop', 'stopAct'); });
        $("#restartAct").click(function(){ performServiceAction('restart', 'restartAct'); });
        $("#statusAct").click(function(){ performServiceAction('status', 'statusAct'); });
        $("#testAct").click(function(){ performServiceAction('test', 'testAct'); });

        // View logs
        $("#logAct").click(function(){
            $("#logAct_progress").addClass("fa fa-spinner fa-pulse");
            ajaxCall(url="/api/dhcpadguardsync/service/logs", sendData={}, callback=function(data,status) {
                $("#logAct_progress").removeClass("fa fa-spinner fa-pulse");
                if (status === "success") {
                    $("#logOutput").text(data.response || data.status);
                } else {
                    $("#logOutput").text("{{ lang._('Error retrieving logs.') }}");
                }
            });
        });
    });
</script>

<div class="alert alert-info hidden" role="alert" id="responseMsg"></div>

<!-- Service Controls -->
<div class="content-box" style="padding-bottom: 1.5em;">
    <div class="col-md-12">
        <h4>{{ lang._('Service Control') }}</h4>
        <button class="btn btn-success" id="startAct" type="button">
            <b>{{ lang._('Start') }}</b> <i id="startAct_progress"></i>
        </button>
        <button class="btn btn-warning" id="stopAct" type="button">
            <b>{{ lang._('Stop') }}</b> <i id="stopAct_progress"></i>
        </button>
        <button class="btn btn-info" id="restartAct" type="button">
            <b>{{ lang._('Restart') }}</b> <i id="restartAct_progress"></i>
        </button>
        <button class="btn btn-default" id="statusAct" type="button">
            <b>{{ lang._('Status') }}</b> <i id="statusAct_progress"></i>
        </button>
        <hr/>
        <pre id="serviceOutput"></pre>
    </div>
</div>

<!-- Main tabs -->
<ul class="nav nav-tabs" role="tablist" id="maintabs">
    <li class="active"><a data-toggle="tab" href="#settings">{{ lang._('Settings') }}</a></li>
    <li><a data-toggle="tab" href="#logs">{{ lang._('Logs') }}</a></li>
</ul>

<div class="tab-content content-box">
    <div id="settings" class="tab-pane fade in active">
        <div class="content-box" style="padding-bottom: 1.5em;">
            {{ partial("layout_partials/base_form",["fields":generalForm,"id":"frm_GeneralSettings"]) }}
            <div class="col-md-12">
                <hr />
                <button class="btn btn-primary" id="saveAct" type="button">
                    <b>{{ lang._('Save') }}</b> <i id="saveAct_progress"></i>
                </button>
                <button class="btn btn-info" id="testAct" type="button">
                    <b>{{ lang._('Test Configuration') }}</b> <i id="testAct_progress"></i>
                </button>
            </div>
        </div>
    </div>
    <div id="logs" class="tab-pane fade">
        <div class="content-box">
            <div class="col-md-12">
                <button class="btn btn-primary" id="logAct" type="button">
                    <b>{{ lang._('View Logs') }}</b> <i id="logAct_progress"></i>
                </button>
                <hr/>
                <pre id="logOutput"></pre>
            </div>
        </div>
    </div>
</div>
