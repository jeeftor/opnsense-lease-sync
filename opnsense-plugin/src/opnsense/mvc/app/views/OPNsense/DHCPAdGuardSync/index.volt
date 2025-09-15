{% extends "layout_partials/base_form.volt" %}
{% block content %}
<div class="alert alert-info hidden" role="alert" id="responseMsg">
</div>

<ul class="nav nav-tabs" role="tablist" id="maintabs">
    <li class="active"><a data-toggle="tab" href="#settings">{{ lang._('Settings') }}</a></li>
    <li><a data-toggle="tab" href="#status">{{ lang._('Status') }}</a></li>
    <li><a data-toggle="tab" href="#logs">{{ lang._('Logs') }}</a></li>
</ul>

<div class="tab-content content-box">
    <div id="settings" class="tab-pane fade in active">
        <div class="content-box" style="padding-bottom: 1.5em;">
            {{ partial("layout_partials/base_form",['fields':formSettings,'id':'frm_settings']) }}
            <div class="col-md-12">
                <hr />
                <button class="btn btn-primary" id="saveAct" type="button"><b>{{ lang._('Save') }}</b> <i id="saveAct_progress"></i></button>
            </div>
        </div>
    </div>
    <div id="status" class="tab-pane fade">
        <div class="content-box">
            <div class="col-md-12">
                <button class="btn btn-primary" id="statusAct" type="button"><b>{{ lang._('Get Status') }}</b> <i id="statusAct_progress"></i></button>
                <hr/>
                <pre id="statusOutput"></pre>
            </div>
        </div>
    </div>
    <div id="logs" class="tab-pane fade">
        <div class="content-box">
            <div class="col-md-12">
                <button class="btn btn-primary" id="logAct" type="button"><b>{{ lang._('View Logs') }}</b> <i id="logAct_progress"></i></button>
                <hr/>
                <pre id="logOutput"></pre>
            </div>
        </div>
    </div>
</div>

<script>
    $(document).ready(function() {
        var data_get_map = {'frm_settings':"/api/dhcpadguardsync/settings/get"};
        mapDataToFormUI(data_get_map).done(function(data){
            formatTokenizersUI();
            $('.selectpicker').selectpicker('refresh');
        });

        // DHCP Server type change handler
        $("#dhcpadguardsync\\.general\\.dhcp_server").change(function(){
            var serverType = $(this).val();
            var leasePathField = $("#dhcpadguardsync\\.general\\.lease_path");
            var leaseFormatField = $("#dhcpadguardsync\\.general\\.lease_format");

            if (serverType === 'isc') {
                leasePathField.val('/var/dhcpd/var/db/dhcpd.leases');
                leaseFormatField.val('isc').trigger('change');
            } else if (serverType === 'dnsmasq') {
                leasePathField.val('/var/db/dnsmasq.leases');
                leaseFormatField.val('dnsmasq').trigger('change');
            }
            // For 'custom', leave fields as-is for manual configuration
        });

        // Save settings
        $("#saveAct").click(function(){
            saveFormToEndpoint(url="/api/dhcpadguardsync/settings/set", formid='frm_settings',callback_ok=function(){
                $("#responseMsg").removeClass("hidden").html("{{ lang._('Settings saved. Please apply changes to activate.') }}");
            });
        });

        // Get status
        $("#statusAct").click(function(){
            $("#statusAct_progress").addClass("fa fa-spinner fa-pulse");
            ajaxCall(url="/api/dhcpadguardsync/service/status", sendData={}, callback=function(data,status) {
                $("#statusAct_progress").removeClass("fa fa-spinner fa-pulse");
                $("#statusOutput").text(data['response']);
            });
        });

        // Get logs
        $("#logAct").click(function(){
            $("#logAct_progress").addClass("fa fa-spinner fa-pulse");
            ajaxCall(url="/api/dhcpadguardsync/service/logs", sendData={}, callback=function(data,status) {
                $("#logAct_progress").removeClass("fa fa-spinner fa-pulse");
                $("#logOutput").text(data['response']);
            });
        });
    });
</script>
{% endblock %}
