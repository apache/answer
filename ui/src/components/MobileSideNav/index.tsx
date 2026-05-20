/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { Offcanvas } from 'react-bootstrap';
import { useLocation } from 'react-router-dom';

import { SideNav, AdminSideNav } from '@/components';

import './index.scss';

const MobileSideNav = ({ show, onHide }) => {
  const { pathname } = useLocation();
  const isAdmin = pathname.includes('/admin');
  return (
    <Offcanvas
      show={show}
      onHide={() => {
        onHide(false);
      }}
      id="mobileSideNav"
      className="px-3 py-4">
      <Offcanvas.Body className="p-0">
        {isAdmin ? <AdminSideNav /> : <SideNav />}
      </Offcanvas.Body>
    </Offcanvas>
  );
};

export default MobileSideNav;
