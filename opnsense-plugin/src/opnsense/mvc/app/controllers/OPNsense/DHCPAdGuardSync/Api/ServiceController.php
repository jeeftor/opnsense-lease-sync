<?php

/**
 *    Copyright (C) 2024 DHCP AdGuard Sync
 *    All rights reserved.
 *
 *    Redistribution and use in source and binary forms, with or without
 *    modification, are permitted provided that the following conditions are met:
 *
 *    1. Redistributions of source code must retain the above copyright notice,
 *       this list of conditions and the following disclaimer.
 *
 *    2. Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 *    THIS SOFTWARE IS PROVIDED ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES,
 *    INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
 *    AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
 *    AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY,
 *    OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 *    SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 *    INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 *    CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 *    ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 *    POSSIBILITY OF SUCH DAMAGE.
 */

namespace OPNsense\DHCPAdGuardSync\Api;

use OPNsense\Base\ApiMutableServiceControllerBase;
use OPNsense\DHCPAdGuardSync\DHCPAdGuardSync;
use OPNsense\Core\Backend;

class ServiceController extends ApiMutableServiceControllerBase
{
    protected static $internalServiceClass = '\OPNsense\DHCPAdGuardSync\DHCPAdGuardSync';
    protected static $internalServiceTemplate = 'OPNsense/DHCPAdGuardSync';
    protected static $internalServiceEnabled = 'general.enabled';
    protected static $internalServiceName = 'dhcpadguardsync';

    /**
     * Additional custom action for testing configuration
     */
    public function testAction()
    {
        $result = array("result" => "failed");
        if ($this->request->isPost()) {
            $backend = new Backend();
            $response = $backend->configdRun("dhcpadguardsync test");
            if (strpos($response, "OK") !== false || strpos($response, "success") !== false) {
                $result['result'] = 'ok';
            }
            $result['response'] = $response;
        }
        return $result;
    }

    /**
     * Get service logs
     */
    public function logsAction()
    {
        $result = array("result" => "failed");
        if ($this->request->isPost()) {
            $backend = new Backend();
            $response = $backend->configdRun("dhcpadguardsync logs");
            $result['response'] = $response;
            $result['result'] = 'ok';
        }
        return $result;
    }
}
