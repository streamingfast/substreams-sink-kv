import React, {useState} from 'react';
import {Breadcrumb, Layout, Menu, theme, DatePicker, Col, Row, Button, Table, Tag, Tooltip, Space, Alert} from 'antd';
import { SearchOutlined,  ClockCircleOutlined } from '@ant-design/icons';
import logo from './assets/logo.png';
import {
    createConnectTransport,
    createPromiseClient,
} from "@bufbuild/connect-web";

// Import service definition that you want to connect to.
import { BlockMeta } from "./pb/proto/service_connectweb";
import {Month, Months} from "./pb/proto/service_pb";

import './App.css'
import {RangePickerProps} from "antd/es/date-picker";

const { Header, Content, Footer } = Layout;
const { RangePicker } = DatePicker;

// The transport defines what type of endpoint we're hitting.
// In our example we'll be communicating with a Connect endpoint.
const transport = createConnectTransport({
    baseUrl: import.meta.env.VITE_API_URL || "http://localhost:7878",
});



// Here we make the client itself, combining the service
// definition with the transport.
const client = createPromiseClient(BlockMeta, transport);

const bufferToHex = (buffer: Uint8Array): string => {
    var s = '', h = '0123456789abcdef';
    (new Uint8Array(buffer)).forEach((v) => { s += h[v >> 4] + h[v & 15]; });
    return s;
}

const App: React.FC = () => {
    const [startDate, setStartDate] = useState<string>()
    const [endDate, setEndDate] = useState<string>()
    const [loading,setLoading]= useState(false);
    const [months,setMonths] = useState<Month[]>([])

    const {
        token: { colorBgContainer },
    } = theme.useToken();



    const onChange: RangePickerProps['onChange'] = (date, dateString) => {
        setStartDate(dateString[0])
        setEndDate(dateString[1])
    };

    const search = async () => {
        setLoading(true)
        const response = await client.getBlockInfo({
            start: startDate,
            end: endDate,
        })
        setMonths(response.months)
        setLoading(false)
    }

    const columns = [
        {
            title: 'Date',
            key: 'date',
            render: (_: any, month: Month) => (
                <>
                    <ClockCircleOutlined style={{marginRight: "10px"}}/>
                    {month.year}-{month.month}
                </>
            ),

        },
        {
            title: 'First Block',
            key: 'first_block',
            render: (_: any, month: Month) => {
                if (!month.firstBlock) {
                    return (<></>)
                }
                return (
                    <>
                        <a href={"https://etherscan.io/block/"+month.firstBlock.number.toString()}>{month.firstBlock.number.toString()}</a>
                        <Tag style={{marginLeft: "10px"}}>0x{bufferToHex(month.firstBlock.hash).substring(1,19)}...</Tag>
                        @ {month.firstBlock.timestamp}
                    </>
                )
            },

        },
        {
            title: 'Last Block',
            key: 'last_block',
            render: (_: any, month: Month) => {
                if (!month.lastBlock) {
                    return (<></>)
                }
                return (
                    <>
                        <a href={"https://etherscan.io/block/"+month.lastBlock.number.toString()}>{month.lastBlock.number.toString()}</a>
                        <Tag style={{marginLeft: "10px"}}>0x{bufferToHex(month.lastBlock.hash).substring(1,19)}...</Tag>
                        @ {month.lastBlock.timestamp}
                    </>
                )
            },

        },
    ];

    return (
        <Layout className="layout">
            <Header>
                <div style={{color: "white"}} >
                    <img style={{height: "60px"}} src={logo} className="App-logo" alt="logo" />
                </div>
                <Menu
                    theme="dark"
                    mode="horizontal"
                    defaultSelectedKeys={['2']}
                />
            </Header>
            <Content style={{ padding: '0 50px' }}>
                <div className="site-layout-content" style={{ background: colorBgContainer, marginTop: "25px" }} >
                    <Row justify={"center"} align={'middle'} >
                        <Col md={24} style={{textAlign: 'center'}}>
                            <h1>ETH BlockMeta powered by connect-web and Substream Sink KV</h1>
                        </Col>
                    </Row>
                    <Row gutter={[16, 16]} style={{marginTop: "25px"}}>
                        <Col><RangePicker picker="month" onChange={onChange}/></Col>
                        <Col>
                            <Button
                                type="primary"
                                icon={<SearchOutlined />}
                                onClick={search}
                                loading={loading}
                                disabled={!(startDate && endDate)}
                            >
                                Search
                            </Button>
                        </Col>
                    </Row>
                    <Table dataSource={months} columns={columns} style={{marginTop: "20px"}}/>
                </div>
            </Content>
            <Footer style={{ textAlign: 'center' }}>StreamingFast ©2023</Footer>
        </Layout>
    );
};

export default App;