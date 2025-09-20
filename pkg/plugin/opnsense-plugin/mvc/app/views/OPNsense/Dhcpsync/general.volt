<script>
    $( document ).ready(function() {
        var data_get_map = {'frm_GeneralSettings':"/api/dhcpsync/general/get"};
        mapDataToFormUI(data_get_map).done(function(data) {
            formatTokenizersUI();
            $('.selectpicker').selectpicker('refresh');

            // Initialize select fields
            $('select').each(function(){
                if ($(this).data('allownew') === undefined) {
                    $(this).selectpicker();
                }
            });

            // Refresh all selectpickers
            $('.selectpicker').selectpicker('refresh');
        });

        $("#reconfigureAct").SimpleActionButton({
            onPreAction: function () {
              const dfObj = new $.Deferred();

              // Special handling for password field
              var passwordField = $("#dhcpsync\\.general\\.AdguardHomePassword");
              var passwordValue = passwordField.val();

              // If validation error is showing, remove it
              passwordField.closest('.form-group').removeClass('has-error');
              passwordField.closest('.form-group').find('.help-block').text('');

              saveFormToEndpoint("/api/dhcpsync/general/set", 'frm_GeneralSettings', function () { dfObj.resolve(); }, true, function () { dfObj.reject(); });
              return dfObj;
            }
        });

        updateServiceControlUI('dhcpsync');

        // Test Connection button handler
        $("#testConnectionBtn").click(function() {
            // Show loading indicator
            $("#testConnectionStatus").html('<i class="fa fa-spinner fa-spin"></i> {{ lang._("Testing connection...") }}');
            console.log("Test Connection button clicked");

            // Save form data first to ensure we're testing with current values
            saveFormToEndpoint("/api/dhcpsync/general/set", 'frm_GeneralSettings', function() {
                console.log("Form data saved, now testing connection");

                // After saving, test the connection
                $.ajax({
                    url: "/api/dhcpsync/general/testConnection",
                    type: "POST",
                    dataType: "json",
                    success: function(data) {
                        console.log("Test connection response received:", data);
                        if (data.status === "ok") {
                            $("#testConnectionStatus").html('<span class="text-success"><i class="fa fa-check"></i> ' + data.message + '</span>');
                        } else {
                            $("#testConnectionStatus").html('<span class="text-danger"><i class="fa fa-times"></i> ' + data.message + '</span>');

                            // Add debug button if debug info is available
                            if (data.debug) {
                                $("#testConnectionStatus").append(' <button class="btn btn-xs btn-info" id="showDebugBtn"><i class="fa fa-bug"></i> Show Debug Info</button>');

                                // Create hidden debug info panel
                                if ($("#debugInfoPanel").length === 0) {
                                    $("#testConnectionStatus").after('<div id="debugInfoPanel" class="panel panel-default" style="display:none; margin-top:10px;">' +
                                        '<div class="panel-heading"><h3 class="panel-title">Debug Information</h3></div>' +
                                        '<div class="panel-body"><pre id="debugInfoContent"></pre></div>' +
                                    '</div>');
                                }

                                // Format debug info
                                var debugContent = 'URL: ' + data.debug.url + '\n' +
                                    'HTTP Code: ' + data.debug.http_code + '\n' +
                                    'Error: ' + (data.debug.error || 'None') + '\n\n' +
                                    'Headers:\n' + data.debug.headers + '\n\n' +
                                    'Response:\n' + data.debug.response_preview;

                                $("#debugInfoContent").text(debugContent);

                                // Show/hide debug panel on button click
                                $("#showDebugBtn").click(function() {
                                    $("#debugInfoPanel").toggle();
                                });
                            }
                        }
                    },
                    error: function(xhr, status, error) {
                        console.log("AJAX error:", status, error);
                        console.log("XHR:", xhr);
                        $("#testConnectionStatus").html('<span class="text-danger"><i class="fa fa-times"></i> Request failed: ' + error + '</span>');
                    }
                });
            });
        });
    });
</script>

<div class="alert alert-info hidden" role="alert" id="responseMsg"></div>
<div class="content-box __mb">
    DHCP AdGuard Sync will sync DHCP leases & ARP table entries with AdGuard Home - so that you can better identifiy clients in the client list.
    {{ partial("layout_partials/base_form",['fields':generalForm,'id':'frm_GeneralSettings'])}}
</div>

<div class="row">
    <div class="col-md-12">
        <button class="btn btn-primary" id="testConnectionBtn" type="button"><i class="fa fa-plug"></i> {{ lang._('Test Adguard Home Connection') }}</button>
        <span id="testConnectionStatus" style="margin-left: 10px;"></span>
    </div>
</div>
<br>

{{ partial('layout_partials/base_apply_button', {'data_endpoint': '/api/dhcpsync/service/reconfigure', 'data_service_widget': 'dhcpsync'}) }}
