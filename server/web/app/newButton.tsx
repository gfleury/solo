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
                    <Dropdown.Item href="#add-instagram-non-official" onClick={() => router.push("/accounts/add")}>Add Instagram (non-official)</Dropdown.Item>
                    <Dropdown.Item href="#add-instagram-official" onClick={() => router.push(GetUrlAPI()+"/provider/Instagram")}>Add Instagram (official)</Dropdown.Item>
                    <Dropdown.Item href="#add-youtube">Add Youtube</Dropdown.Item>
                </Dropdown.Menu>
            </Dropdown>
        </Nav>
    );
}