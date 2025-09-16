{#
 # Copyright (c) 2024 DHCP AdGuard Sync
 # All rights reserved.
 #
 # Redistribution and use in source and binary forms, with or without
 # modification, are permitted provided that the following conditions are met:
 #
 # 1. Redistributions of source code must retain the above copyright notice,
 #    this list of conditions and the following disclaimer.
 #
 # 2. Redistributions in binary form must reproduce the above copyright
 #    notice, this list of conditions and the following disclaimer in the
 #    documentation and/or other materials provided with the distribution.
 #
 # THIS SOFTWARE IS PROVIDED ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES,
 # INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
 # AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
 # AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
 # OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 # SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 # INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 # CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 # ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 # POSSIBILITY OF SUCH DAMAGE.
 #}

<script>
    $( document ).ready(function() {
        var data_get_map = {'frm_dhcpadguardsync':"/api/dhcpadguardsync/settings/get"};
        mapDataToFormUI(data_get_map).done(function(data){
            formatTokenizersUI();
            $('.selectpicker').selectpicker('refresh');
        });

        // link save button to API set action
        $("#saveAct").click(function(){
            saveFormToEndpoint(url="/api/dhcpadguardsync/settings/set", formid='frm_dhcpadguardsync',callback_ok=function(){
                $("#responseMsg").removeClass("hidden");
                $("#responseMsg").html("{{ lang._('Settings saved') }}");
                // auto-reconfigure service after save
                ajaxCall(url="/api/dhcpadguardsync/service/reconfigure", sendData={}, callback=function(data,status) {
                    if (status === "success" && data.status === 'ok') {
                        $("#responseMsg").html("{{ lang._('Settings saved and service reconfigured') }}");
                    }
                });
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
                    $("#serviceOutput").text("Error performing " + action);
                }
            });
        }

        $("#startAct").click(function(){ performServiceAction('start', 'startAct'); });
        $("#stopAct").click(function(){ performServiceAction('stop', 'stopAct'); });
        $("#restartAct").click(function(){ performServiceAction('restart', 'restartAct'); });
        $("#statusAct").click(function(){ performServiceAction('status', 'statusAct'); });

        // Test configuration
        $("#testAct").click(function(){
            $("#testAct_progress").addClass("fa fa-spinner fa-pulse");
            ajaxCall(url="/api/dhcpadguardsync/service/test", sendData={}, callback=function(data,status) {
                $("#testAct_progress").removeClass("fa fa-spinner fa-pulse");
                if (status === "success") {
                    $("#serviceOutput").text(data.response || data.status);
                } else {
                    $("#serviceOutput").text("Error testing configuration");
                }
            });
        });

        // View logs
        $("#logAct").click(function(){
            $("#logAct_progress").addClass("fa fa-spinner fa-pulse");
            ajaxCall(url="/api/dhcpadguardsync/service/logs", sendData={}, callback=function(data,status) {
                $("#logAct_progress").removeClass("fa fa-spinner fa-pulse");
                if (status === "success") {
                    $("#logOutput").text(data.response || data.status);
                } else {
                    $("#logOutput").text("Error retrieving logs");
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
            {{ partial("layout_partials/base_form",["fields":generalForm,"id":"frm_dhcpadguardsync"]) }}
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
