import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';

export default function SearchBar() {
    return (
        <Form>
            <InputGroup className="mb-3">
                <InputGroup.Text id="search-bar"><FontAwesomeIcon icon={faSearch}/></InputGroup.Text>
                <Form.Control
                    placeholder="Search"
                    aria-label="Search"
                    aria-describedby="search-bar"
                />
            </InputGroup>
        </Form>
    )
}