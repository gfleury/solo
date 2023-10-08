'use client'

import { Get, Delete, HandleFailures } from '../../api-client'
import { Network, NetworkNode } from '../../api-client/models'
import { useSWRConfig } from 'swr'
import { useRouter } from 'next/navigation'
import Loading from '../../loading'
import SearchBar from './search'
import { Stack } from 'react-bootstrap'
import Table from 'react-bootstrap/Table';
import ButtonGroup from 'react-bootstrap/ButtonGroup'
import Button from 'react-bootstrap/Button'
import { faRemove } from '@fortawesome/free-solid-svg-icons'
import { OneOrBothBySize } from '../../utils'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faApple, faLinux } from '@fortawesome/free-brands-svg-icons'

export default function Network({ params }: { params: { networkID: string } }) {
  const { mutate } = useSWRConfig()
  const router = useRouter()
  const network = HandleFailures(Get(`/network/${params.networkID}`), router)
  const nodeDelete = Delete("/node/")

  function getOSIcon(name: string) {
    if (name == "linux") {
      return <FontAwesomeIcon icon={faLinux} style={{ color: "#f3eff0", }} />
    }
    if (name == "darwin") {
      return <FontAwesomeIcon icon={faApple} style={{ color: "#f3eff0", }} />
    }
  }

  if (network.isLoading) return (Loading())

  let networkCast: Network = network.data

  console.log(networkCast)

  return (
    <Stack className="my-2" gap={2}>
      <div>
        <SearchBar />
      </div>
      <div>
        <Table hover align="center" responsive>
          <thead>
            <tr>
              <th>Hostname</th>
              <th>Arch</th>
              <th>IP address</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {networkCast.nodes?.map((node: NetworkNode) => (
              <tr key={node.ID}>
                <td>
                  {getOSIcon(node.OS!)}&nbsp;
                  {node.Hostname}
                </td>
                <td>
                  {node.Arch}
                </td>
                <td>
                  {node.IP}
                </td>
                <td>
                  <ButtonGroup size='sm'>
                    <Button variant="primary" disabled={nodeDelete.isMutating}>
                      {OneOrBothBySize("Delete", <FontAwesomeIcon icon={faRemove} />)}
                    </Button>
                    {/* <Edit nodeID={node.ID !== undefined ? node.ID : -1} /> */}
                  </ButtonGroup>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      </div>
    </Stack>
  )
}