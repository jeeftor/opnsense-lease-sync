<script>
    $(document).ready(function() {
        var data_get_map = {'frm_GeneralSettings':"/api/dhcpadguardsync/settings/get"};
        mapDataToFormUI(data_get_map).done(function(data){
            formatTokenizersUI();
            $('.selectpicker').selectpicker('refresh');
        });

        $("#saveAct").click(function(){
            saveFormToEndpoint(url="/api/dhcpadguardsync/settings/set",formid='frm_GeneralSettings',callback_ok=function(){
                $("#saveAct").prop('disabled',true);
                ajaxCall(url="/api/dhcpadguardsync/service/restart",sendData={},callback=function(data,status){
                    $("#saveAct").prop('disabled',false);
                });
            });
        });
    });
</script>

<div class="tab-content content-box">
    <div id="general" class="tab-pane fade in active">
        <div class="content-box" style="padding-bottom: 1.5em;">
            {{ partial("layout_partials/base_form",['fields':generalForm,'id':'frm_GeneralSettings'])}}
            <div class="col-md-12">
                <hr />
                <button class="btn btn-primary" id="saveAct" type="button"><b>{{ lang._('Save') }}</b> <i id="saveAct_progress"></i></button>
            </div>
        </div>
    </div>
</div>