{#
    # Copyright (c) 2024 DHCP AdGuard Sync
    # All rights reserved.
    #}

{% extends "layout_partials/base.volt" %}

{% block content %}

<script>
       $( document ).ready(function() {
           var data_get_map = {'frm_settings':"/api/dhcpadguardsync/settings/get"};
           mapDataToFormUI(data_get_map).done(function(data){
               formatTokenizersUI();
               $('.selectpicker').selectpicker('refresh');
           });

           // Auto-save and restart on settings change
           $("#saveAct").click(function(){
               $("#saveAct_progress").addClass("fa fa-spinner fa-pulse");
               saveFormToEndpoint(url="/api/dhcpadguardsync/settings/set", formid='frm_settings',callback_ok=function(){
                   $("#responseMsg").removeClass("hidden").html("{{ lang._('Settings saved and service restarted.') }}");
                   $("#saveAct_progress").removeClass("fa fa-spinner fa-pulse");
                   // Auto-restart service after save
                   ajaxCall(url="/api/dhcpadguardsync/service/restart", sendData={}, callback=function(data,status) {
                       $("#serviceOutput").text("Service restarted: " + data['response']);
                   });
               }, callback_fail=function(){
                   $("#saveAct_progress").removeClass("fa fa-spinner fa-pulse");
               });
           });

           // Service control buttons
           $("#startAct").click(function(){
               $("#startAct_progress").addClass("fa fa-spinner fa-pulse");
               ajaxCall(url="/api/dhcpadguardsync/service/start", sendData={}, callback=function(data,status) {
                   $("#startAct_progress").removeClass("fa fa-spinner fa-pulse");
                   $("#serviceOutput").text(data['response']);
               });
           });

           $("#stopAct").click(function(){
               $("#stopAct_progress").addClass("fa fa-spinner fa-pulse");
               ajaxCall(url="/api/dhcpadguardsync/service/stop", sendData={}, callback=function(data,status) {
                   $("#stopAct_progress").removeClass("fa fa-spinner fa-pulse");
                   $("#serviceOutput").text(data['response']);
               });
           });

           $("#restartAct").click(function(){
               $("#restartAct_progress").addClass("fa fa-spinner fa-pulse");
               ajaxCall(url="/api/dhcpadguardsync/service/restart", sendData={}, callback=function(data,status) {
                   $("#restartAct_progress").removeClass("fa fa-spinner fa-pulse");
                   $("#serviceOutput").text(data['response']);
               });
           });


           // Test configuration
           $("#testAct").click(function(){
               $("#testAct_progress").addClass("fa fa-spinner fa-pulse");
               ajaxCall(url="/api/dhcpadguardsync/service/test", sendData={}, callback=function(data,status) {
                   $("#testAct_progress").removeClass("fa fa-spinner fa-pulse");
                   $("#serviceOutput").text(data['response']);
               });
           });

           // Get status
           $("#statusAct").click(function(){
               $("#statusAct_progress").addClass("fa fa-spinner fa-pulse");
               ajaxCall(url="/api/dhcpadguardsync/service/status", sendData={}, callback=function(data,status) {
                   $("#statusAct_progress").removeClass("fa fa-spinner fa-pulse");
                   $("#serviceOutput").text(data['response']);
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

   <div class="alert alert-info hidden" role="alert" id="responseMsg">
   </div>

   <!-- Service Controls -->
   <div class="content-box">
       <div class="col-md-12">
           <h4>{{ lang._('Service Control') }}</h4>
           <button class="btn btn-success" id="startAct" type="button"><b>{{ lang._('Start') }}</b> <i id="startAct_progress"></i></button>
           <button class="btn btn-warning" id="stopAct" type="button"><b>{{ lang._('Stop') }}</b> <i id="stopAct_progress"></i></button>
           <button class="btn btn-info" id="restartAct" type="button"><b>{{ lang._('Restart') }}</b> <i id="restartAct_progress"></i></button>
           <button class="btn btn-default" id="statusAct" type="button"><b>{{ lang._('Status') }}</b> <i id="statusAct_progress"></i></button>
           <hr/>
           <pre id="serviceOutput"></pre>
       </div>
   </div>

   <ul class="nav nav-tabs" role="tablist" id="maintabs">
       <li class="active"><a data-toggle="tab" href="#settings">{{ lang._('Configuration') }}</a></li>
       <li><a data-toggle="tab" href="#logs">{{ lang._('Logs') }}</a></li>
   </ul>

   <div class="tab-content content-box">
       <div id="settings" class="tab-pane fade in active">
           <div class="content-box" style="padding-bottom: 1.5em;">
               {{ partial("layout_partials/base_form",["fields":generalForm,"id":"frm_settings"]) }}
               <div class="col-md-12">
                   <hr />
                   <button class="btn btn-primary" id="saveAct" type="button"><b>{{ lang._('Save') }}</b> <i id="saveAct_progress"></i></button>
                   <button class="btn btn-info" id="testAct" type="button"><b>{{ lang._('Test Configuration') }}</b> <i id="testAct_progress"></i></button>
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

{% endblock %}
