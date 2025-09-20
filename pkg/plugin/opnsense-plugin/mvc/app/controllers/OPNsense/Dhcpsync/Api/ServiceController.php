<?php

/*
 * Copyright (C) 2021 Michael Muenz <michael.muenz@max-it.de>
 * All rights reserved.
 */

namespace OPNsense\Dhcpsync\Api;

use OPNsense\Base\ApiMutableServiceControllerBase;

/**
 * ServiceController for the dhcpsync plugin.
 *
 * This controller handles the start, stop, restart, and status actions
 * for the main dhcpsync service.
 */
class ServiceController extends ApiMutableServiceControllerBase
{
    protected static $internalServiceClass = '\\OPNsense\\Dhcpsync\\General';
    protected static $internalServiceTemplate = 'OPNsense/Dhcpsync';
    protected static $internalServiceEnabled = 'enabled';
    protected static $internalServiceName = 'dhcpsync';
}
