import { Dropdown, Nav } from 'react-bootstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPlus } from '@fortawesome/free-solid-svg-icons'
import { useRouter } from 'next/navigation'
import {GetUrlAPI } from './api-client'

export default function NewButton() {
    const router = useRouter()

    return (
        <Nav>
            <Dropdown>
                <Dropdown.Toggle variant="primary" id="dropdown-basic">
                    <FontAwesomeIcon icon={faPlus} /> New
                </Dropdown.Toggle>

                <Dropdown.Menu>
                    <Dropdown.Item href="#add-network" onClick={() => router.push("/networks/add")}>Add Network</Dropdown.Item>
                    <Dropdown.Item href="#add-node" onClick={() => router.push("/nodes/add")}>Add Node</Dropdown.Item>
                </Dropdown.Menu>
            </Dropdown>
        </Nav>
    );
}